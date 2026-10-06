package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/claude"
	"github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/documents"
	gh "github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/github"
	"github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/workflow"
)

type eventPayload struct {
	Action string `json:"action"`
	Issue  struct {
		Number int    `json:"number"`
		State  string `json:"state"`
	} `json:"issue"`
	Label struct {
		Name string `json:"name"`
	} `json:"label"`
	PullRequest struct {
		Merged bool `json:"merged"`
		Head   struct {
			Ref string `json:"ref"`
		} `json:"head"`
	} `json:"pull_request"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	issueNumber := flag.Int("issue", 0, "GitHub VI issue number (local live run)")
	eventPath := flag.String("event", os.Getenv("GITHUB_EVENT_PATH"), "GitHub Actions event JSON path")
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	var event eventPayload
	if *eventPath != "" {
		data, err := os.ReadFile(*eventPath)
		if err != nil {
			return fmt.Errorf("read GitHub event: %w", err)
		}
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("decode GitHub event: %w", err)
		}
	}
	if *issueNumber == 0 {
		if os.Getenv("GITHUB_EVENT_NAME") != "pull_request" {
			*issueNumber = event.Issue.Number
		}
	}

	client, err := gh.NewFromEnvironment()
	if err != nil {
		return err
	}
	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}
	if os.Getenv("GITHUB_EVENT_NAME") == "pull_request" {
		if event.Action == "closed" && event.PullRequest.Merged {
			if issueNumber, ok := issueFromDocsBranch(event.PullRequest.Head.Ref); ok {
				if err := client.SetWorkflowLabel(context.Background(), issueNumber, "vi:done"); err != nil {
					return fmt.Errorf("mark VI done after docs PR merge: %w", err)
				}
				if err := client.AddIssueComment(context.Background(), issueNumber, "The documentation PR was merged. This VI is now labeled `vi:done`."); err != nil {
					return fmt.Errorf("comment on resolved VI: %w", err)
				}
			}
		}
		return nil
	}
	if *issueNumber <= 0 {
		return fmt.Errorf("provide --issue or run from a GitHub issue event")
	}

	action := event.Action
	label := event.Label.Name
	if os.Getenv("GITHUB_EVENT_NAME") == "workflow_dispatch" {
		action = "labeled"
		label = "vi:ready-for-docs"
	}
	if action == "" && *eventPath == "" {
		action = "labeled"
		label = "vi:ready-for-docs"
	}
	if action == "closed" {
		executor := workflow.NewExecutor(client, documents.Mapping{}, *issueNumber, absoluteRoot, "")
		if err := executor.HandlePrematureClosure(context.Background()); err != nil {
			return fmt.Errorf("recover prematurely closed VI: %w", err)
		}
		return nil
	}
	if action != "labeled" || label != "vi:ready-for-docs" {
		return nil
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	authToken := os.Getenv("ANTHROPIC_AUTH_TOKEN")
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	model := os.Getenv("ANTHROPIC_MODEL")
	if (apiKey == "" && authToken == "") || model == "" {
		return fmt.Errorf("live agent run requires ANTHROPIC_MODEL and ANTHROPIC_API_KEY or ANTHROPIC_AUTH_TOKEN")
	}
	mappingFile, err := os.Open(filepath.Join(absoluteRoot, "mappings", "docs-map.yaml"))
	if err != nil {
		return fmt.Errorf("open docs mapping: %w", err)
	}
	mapping, mappingErr := documents.LoadMapping(mappingFile)
	closeErr := mappingFile.Close()
	if mappingErr != nil {
		return mappingErr
	}
	if closeErr != nil {
		return closeErr
	}

	executor := workflow.NewExecutor(client, mapping, *issueNumber, absoluteRoot, model)
	if baseURL != "" {
		parsed, parseErr := url.Parse(baseURL)
		if parseErr != nil || parsed.Host == "" {
			return fmt.Errorf("ANTHROPIC_BASE_URL is not a valid URL")
		}
		executor.Gateway = parsed.Host
	}
	claudeClient := claude.Client{APIKey: apiKey, AuthToken: authToken, BaseURL: baseURL, Model: model}
	result, err := claudeClient.Run(context.Background(), workflow.SystemPrompt(), workflow.UserPrompt(*issueNumber), executor)
	if err != nil {
		if traceErr := saveToolTrace(absoluteRoot, *issueNumber, result); traceErr != nil {
			fmt.Fprintf(os.Stderr, "save partial tool trace: %v\n", traceErr)
		}
		appendToolSummary(result)
		return fmt.Errorf("resolve VI with Claude: %w", err)
	}
	if !executor.Submitted {
		return fmt.Errorf("Claude completed without submitting a Doc Contract")
	}
	if err := saveToolTrace(absoluteRoot, *issueNumber, result); err != nil {
		return err
	}
	appendToolSummary(result)
	return nil
}

func issueFromDocsBranch(branch string) (int, bool) {
	const prefix = "docs/vi-"
	if !strings.HasPrefix(branch, prefix) {
		return 0, false
	}
	number, err := strconv.Atoi(strings.TrimPrefix(branch, prefix))
	return number, err == nil && number > 0
}

func saveToolTrace(root string, issueNumber int, result claude.RunResult) error {
	if err := os.MkdirAll(filepath.Join(root, "artifacts"), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode agent trace: %w", err)
	}
	path := filepath.Join(root, "artifacts", fmt.Sprintf("tool-trace-%d.json", issueNumber))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("save agent trace: %w", err)
	}
	return nil
}

func appendToolSummary(result claude.RunResult) {
	path := os.Getenv("GITHUB_STEP_SUMMARY")
	if path == "" {
		return
	}
	var summary strings.Builder
	summary.WriteString("\n### Live Claude tool calls\n\n")
	for index, call := range result.ToolCalls {
		fmt.Fprintf(&summary, "%d. `%s`\n", index+1, call.Name)
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.WriteString(summary.String())
}
