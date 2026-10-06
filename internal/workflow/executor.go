package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/claude"
	"github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/documents"
	gh "github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/github"
)

const maxEvidenceChars = 12000

var issueNumberPattern = regexp.MustCompile(`#?([0-9]+)`)

type Executor struct {
	GitHub      *gh.Client
	Mapping     documents.Mapping
	IssueNumber int
	RepoRoot    string
	Model       string
	Gateway     string // host of the Claude gateway, empty for api.anthropic.com
	Issue       gh.Issue
	Pull        gh.PullRequest
	PullFiles   []gh.PullRequestFile
	Targets     []documents.DocumentTarget
	Changed     []string
	Evidence    map[string]documents.Evidence
	ContextRead bool
	Submitted   bool
	Decision    documents.Decision
	DocsPRURL   string
	measurement ContextMeasurement
}

type ContextMeasurement struct {
	RepositoryFiles int     `json:"repository_files"`
	CandidateBytes  int64   `json:"candidate_bytes_estimate"`
	SelectedFiles   int     `json:"selected_files"`
	SelectedChars   int     `json:"selected_characters"`
	ReductionPct    float64 `json:"reduction_percent"`
}

type inspection struct {
	Issue           gh.Issue             `json:"issue"`
	Implementation  gh.PullRequest       `json:"implementation_pr"`
	ChangedFiles    []gh.PullRequestFile `json:"changed_files"`
	ChangedConcepts []string             `json:"changed_concepts"`
}

type documentContext struct {
	SelectedDocuments []documents.DocumentTarget `json:"selected_documents"`
	SourceFiles       map[string]string          `json:"source_files"`
	ExistingDocs      map[string]string          `json:"existing_docs"`
	Templates         map[string]string          `json:"templates"`
	Evidence          []documents.Evidence       `json:"evidence"`
	Measurement       ContextMeasurement         `json:"context_selection"`
}

func NewExecutor(client *gh.Client, mapping documents.Mapping, issueNumber int, repoRoot, model string) *Executor {
	return &Executor{
		GitHub:      client,
		Mapping:     mapping,
		IssueNumber: issueNumber,
		RepoRoot:    repoRoot,
		Model:       model,
		Evidence:    make(map[string]documents.Evidence),
	}
}

func (executor *Executor) Definitions() []claude.ToolDefinition {
	return []claude.ToolDefinition{
		{
			Name:        "inspect_vi",
			Description: "Read the VI issue and its linked merged implementation PR, including changed file patches. Call this first.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`),
		},
		{
			Name:        "read_documentation_context",
			Description: "Select mapped documents and read only the relevant product source, existing docs, and audience templates. Call after inspect_vi.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`),
		},
		{
			Name:        "submit_resolution",
			Description: "Submit the evidence-grounded Doc Contract after inspecting the VI and mapped context. This validates the proposal, then either blocks the VI, marks no-impact complete, or opens a human-review docs PR.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"contract":{"type":"object","description":"Doc Contract. Decisions: propose_docs, needs_clarification, no_docs_impact. Every claim and conflict must cite evidence IDs returned by the tools. Each affected document must use the exact mapped path, audience, type, and template, and include the complete Markdown page with all template headings. provider, work_item, and change are filled in by the system; omit them or leave them empty.","properties":{"version":{"type":"string","enum":["1"]},"decision":{"type":"string","enum":["propose_docs","needs_clarification","no_docs_impact"]},"evidence":{"type":"array","description":"Evidence objects copied from tool results; only the id is required.","items":{"type":"object","properties":{"id":{"type":"string"}},"required":["id"]}},"claims":{"type":"array","items":{"type":"object","properties":{"text":{"type":"string"},"audience":{"type":"string","enum":["customer","internal-developer"]},"evidence_ids":{"type":"array","minItems":1,"items":{"type":"string"}}},"required":["text","audience","evidence_ids"]}},"affected_documents":{"type":"array","items":{"type":"object","properties":{"path":{"type":"string"},"audience":{"type":"string"},"type":{"type":"string"},"template":{"type":"string"},"markdown":{"type":"string","minLength":1}},"required":["path","audience","type","template","markdown"]}},"conflicts":{"type":"array","items":{"type":"object","properties":{"description":{"type":"string"},"evidence_ids":{"type":"array","items":{"type":"string"}},"blocks_resolution":{"type":"boolean"}},"required":["description","evidence_ids","blocks_resolution"]}},"missing_information":{"type":"array","items":{"type":"string"}}},"required":["version","decision","evidence","claims","affected_documents","conflicts","missing_information"]}},"required":["contract"],"additionalProperties":false}`),
		},
	}
}

// Finished reports that submit_resolution succeeded, so the run can end.
func (executor *Executor) Finished() bool { return executor.Submitted }

func (executor *Executor) Execute(ctx context.Context, name string, input json.RawMessage) (any, error) {
	switch name {
	case "inspect_vi":
		return executor.inspectVI(ctx)
	case "read_documentation_context":
		return executor.readDocumentationContext(ctx)
	case "submit_resolution":
		return executor.submitResolution(ctx, input)
	default:
		return nil, fmt.Errorf("unknown tool %q", name)
	}
}

func SystemPrompt() string {
	return `You are the live documentation-resolution agent for a product Value Increment (VI). Use the tools to inspect the VI and merged change before deciding. Treat checked-in product code and configuration as authoritative for current product behavior, settings, and defaults. The VI provides intent, audience, and business context, not proof of shipped behavior.

Compare the VI with the implementation and existing documentation. If the VI contradicts product behavior or required documentation cannot be accurate with available evidence, submit a needs_clarification contract with exact evidence and a concise question. Do not create a docs PR in that case. If there is no mapped documentation impact, submit no_docs_impact. Otherwise propose both mapped customer and internal developer documents, keep their audiences separate, follow every required template heading, cite only evidence IDs returned by tools, and never invent unsupported facts. Include missing information and non-blocking conflicts in the contract.

Call inspect_vi, then read_documentation_context, then submit_resolution. Do not claim that a decision or GitHub change succeeded until the submit_resolution tool result confirms it. Keep the user-facing final response concise and name the resulting VI status and docs PR when present.`
}

func UserPrompt(issueNumber int) string {
	return fmt.Sprintf("Resolve documentation for GitHub VI issue #%d. Inspect its linked implementation PR and determine whether it is ready for docs review, needs clarification, or has no documentation impact. Return the complete Doc Contract through submit_resolution.", issueNumber)
}

func (executor *Executor) inspectVI(ctx context.Context) (inspection, error) {
	issue, err := executor.GitHub.GetIssue(ctx, executor.IssueNumber)
	if err != nil {
		return inspection{}, err
	}
	if !hasLabel(issue.Labels, "vi:ready-for-docs") {
		return inspection{}, fmt.Errorf("issue #%d is not labeled vi:ready-for-docs", issue.Number)
	}
	pullNumber, err := implementationPRNumber(issue.Body)
	if err != nil {
		return inspection{}, err
	}
	pull, err := executor.GitHub.GetPullRequest(ctx, pullNumber)
	if err != nil {
		return inspection{}, err
	}
	if pull.MergedAt == nil || pull.MergeCommitSHA == "" {
		return inspection{}, fmt.Errorf("implementation PR #%d must be merged before documentation resolution", pullNumber)
	}
	files, err := executor.GitHub.GetPullRequestFiles(ctx, pullNumber)
	if err != nil {
		return inspection{}, err
	}
	executor.Issue = issue
	executor.Pull = pull
	executor.PullFiles = files
	executor.Changed = changedConcepts(files, executor.Mapping)
	toolFiles := diffContext(files, executor.Changed)
	executor.addEvidence(documents.Evidence{
		ID:     fmt.Sprintf("VI-%d", issue.Number),
		Type:   "VI",
		Source: issue.HTMLURL,
		Detail: truncate(issue.Title+"\n"+issue.Body, maxEvidenceChars),
	})
	executor.addEvidence(documents.Evidence{
		ID:     fmt.Sprintf("PR-%d", pull.Number),
		Type:   "CODE",
		Source: pull.HTMLURL,
		Detail: truncate(pull.Title+"\n"+pull.Body, maxEvidenceChars),
	})

	for _, file := range files {
		if file.Patch == "" || !containsPatch(toolFiles, file.Filename) {
			continue
		}
		id := "DIFF-" + safeID(file.Filename)
		executor.addEvidence(documents.Evidence{ID: id, Type: "CODE", Source: pull.HTMLURL + "/files", Detail: truncate(file.Filename+"\n"+file.Patch, maxEvidenceChars)})
	}
	return inspection{Issue: issue, Implementation: pull, ChangedFiles: toolFiles, ChangedConcepts: executor.Changed}, nil
}

func (executor *Executor) readDocumentationContext(ctx context.Context) (documentContext, error) {
	if executor.Issue.Number == 0 || executor.Pull.MergeCommitSHA == "" {
		return documentContext{}, fmt.Errorf("inspect_vi must be called before read_documentation_context")
	}
	targets := executor.Mapping.TargetsForConcepts(executor.Changed)
	executor.Targets = targets
	bundle := documentContext{
		SelectedDocuments: targets,
		SourceFiles:       make(map[string]string),
		ExistingDocs:      make(map[string]string),
		Templates:         make(map[string]string),
	}
	if len(targets) == 0 {
		bundle.Evidence = executor.evidenceList()
		executor.ContextRead = true
		bundle.Measurement = executor.contextMeasurement(nil)
		executor.measurement = bundle.Measurement
		return bundle, nil
	}

	selectedPaths := make(map[string]struct{})
	sourcePaths := []string{
		"product/process_monitoring/config.yaml",
		"product/process_monitoring/process_monitor.go",
		"product/process_monitoring/process_monitor_test.go",
	}
	for _, path := range sourcePaths {
		content, _, err := executor.GitHub.GetFile(ctx, path, executor.Pull.MergeCommitSHA)
		if err != nil {
			return documentContext{}, fmt.Errorf("read source evidence %s: %w", path, err)
		}
		text := truncate(string(content), maxEvidenceChars)
		bundle.SourceFiles[path] = text
		selectedPaths[path] = struct{}{}
		kind := "CODE"
		if strings.HasSuffix(path, "_test.go") {
			kind = "TEST"
		}
		if strings.HasSuffix(path, ".yaml") {
			kind = "SCHEMA"
		}
		executor.addEvidence(documents.Evidence{ID: "SRC-" + safeID(path), Type: kind, Source: path + "@" + executor.Pull.MergeCommitSHA, Detail: text})
	}

	for _, target := range targets {
		content, _, err := executor.GitHub.GetFile(ctx, target.Path, executor.Pull.MergeCommitSHA)
		if err != nil {
			return documentContext{}, fmt.Errorf("read selected documentation %s: %w", target.Path, err)
		}
		document := string(content)
		bundle.ExistingDocs[target.Path] = document
		selectedPaths[target.Path] = struct{}{}
		executor.addEvidence(documents.Evidence{ID: "DOC-" + safeID(target.Path), Type: "EXISTING_DOCS", Source: target.Path + "@" + executor.Pull.MergeCommitSHA, Detail: truncate(document, maxEvidenceChars)})
		if _, exists := bundle.Templates[target.Template]; !exists {
			template, _, err := executor.GitHub.GetFile(ctx, target.Template, executor.Pull.MergeCommitSHA)
			if err != nil {
				return documentContext{}, fmt.Errorf("read template %s: %w", target.Template, err)
			}
			bundle.Templates[target.Template] = string(template)
			selectedPaths[target.Template] = struct{}{}
			executor.addEvidence(documents.Evidence{ID: "TPL-" + safeID(target.Template), Type: "EXISTING_DOCS", Source: target.Template + "@" + executor.Pull.MergeCommitSHA, Detail: truncate(string(template), maxEvidenceChars)})
		}
	}

	bundle.Evidence = executor.evidenceList()
	bundle.Measurement = executor.contextMeasurement(selectedPaths)
	executor.measurement = bundle.Measurement
	executor.ContextRead = true
	return bundle, nil
}

func (executor *Executor) submitResolution(ctx context.Context, input json.RawMessage) (any, error) {
	if !executor.ContextRead {
		return nil, fmt.Errorf("read_documentation_context must be called before submit_resolution")
	}
	var payload struct {
		Contract documents.Contract `json:"contract"`
	}
	if err := json.Unmarshal(input, &payload); err != nil {
		return nil, fmt.Errorf("decode Doc Contract: %w", err)
	}
	contract := payload.Contract
	contract.Provider = documents.Provider{Name: "Anthropic Claude", Model: executor.Model}
	contract.WorkItem = issueWorkItem(executor.Issue)
	contract.Change.Feature = "process_monitoring"
	contract.Change.DocumentationAudiences = audiences(executor.Targets)
	contract.Change.CustomerVisible = containsAudience(executor.Targets, "customer")
	contract.Evidence = executor.canonicalEvidence(contract.Evidence)
	for index, proposal := range contract.AffectedDocuments {
		for _, target := range executor.Targets {
			if target.Path == proposal.Path {
				contract.AffectedDocuments[index].Reason = target.Reason
			}
		}
	}
	if contract.Evidence == nil {
		return nil, fmt.Errorf("contract must reference evidence returned by tools")
	}
	if err := documents.ValidateContract(contract, executor.Mapping, executor.RepoRoot); err != nil {
		return map[string]any{"error": err.Error(), "next_action": "revise the contract using the tool evidence"}, err
	}
	if err := validateSettingNames(contract.Change.Settings, executor.Mapping); err != nil {
		return nil, err
	}
	if contract.Decision == documents.DecisionNoDocsImpact && len(executor.Targets) > 0 {
		return nil, fmt.Errorf("no_docs_impact is invalid because the docs mapping selected targets")
	}
	if contract.Decision == documents.DecisionProposeDocs && len(executor.Targets) == 0 {
		return nil, fmt.Errorf("propose_docs is invalid because no docs targets were selected")
	}
	if contract.Decision == documents.DecisionNeedsClarification && len(contract.Conflicts) == 0 && len(contract.MissingInformation) == 0 {
		return nil, fmt.Errorf("needs_clarification requires a conflict or missing-information item")
	}
	for _, conflict := range contract.Conflicts {
		if conflict.BlocksResolution && contract.Decision != documents.DecisionNeedsClarification {
			return nil, fmt.Errorf("blocking conflict requires needs_clarification decision")
		}
	}
	if contract.Decision == documents.DecisionProposeDocs {
		if err := requireAllSelectedDocuments(contract.AffectedDocuments, executor.Targets); err != nil {
			return nil, err
		}
	}
	if err := executor.saveContract(contract); err != nil {
		return nil, err
	}
	executor.Decision = contract.Decision
	executor.Submitted = true

	switch contract.Decision {
	case documents.DecisionNeedsClarification:
		comment := resolutionComment(contract)
		if err := executor.GitHub.SetWorkflowLabel(ctx, executor.Issue.Number, "vi:needs-clarification"); err != nil {
			return nil, publishError(err)
		}
		if err := executor.GitHub.AddIssueComment(ctx, executor.Issue.Number, comment); err != nil {
			return nil, publishError(err)
		}
		executor.appendSummary("Needs clarification", contract, "")
		return map[string]any{"status": "vi:needs-clarification", "message": "VI blocked with evidence and follow-up questions"}, nil
	case documents.DecisionNoDocsImpact:
		if err := executor.GitHub.SetWorkflowLabel(ctx, executor.Issue.Number, "vi:done"); err != nil {
			return nil, publishError(err)
		}
		if err := executor.GitHub.AddIssueComment(ctx, executor.Issue.Number, "No mapped documentation impact was found. The VI can be resolved without a docs PR.\n\nSee the workflow artifact for the Doc Contract and evidence."); err != nil {
			return nil, publishError(err)
		}
		executor.appendSummary("No documentation impact", contract, "")
		return map[string]any{"status": "vi:done", "message": "No docs PR was needed"}, nil
	case documents.DecisionProposeDocs:
		files := make(map[string]string, len(contract.AffectedDocuments))
		for _, proposal := range contract.AffectedDocuments {
			files[proposal.Path] = proposal.Markdown
		}
		pull, err := executor.GitHub.CreateDocumentationPR(ctx, executor.Issue.Number, fmt.Sprintf("docs: update process monitoring for VI #%d", executor.Issue.Number), docsPRBody(contract, executor.Pull), files)
		if err != nil {
			return nil, publishError(err)
		}
		if err := executor.GitHub.SetWorkflowLabel(ctx, executor.Issue.Number, "vi:docs-review"); err != nil {
			return nil, publishError(err)
		}
		executor.DocsPRURL = pull.HTMLURL
		if err := executor.GitHub.AddIssueComment(ctx, executor.Issue.Number, fmt.Sprintf("Documentation proposal is ready for human review: %s\n\nThe VI is now labeled `vi:docs-review`. The documentation PR will not be merged automatically.", pull.HTMLURL)); err != nil {
			return nil, publishError(err)
		}
		executor.appendSummary("Documentation proposal", contract, pull.HTMLURL)
		return map[string]any{"status": "vi:docs-review", "docs_pr": pull.HTMLURL}, nil
	default:
		return nil, fmt.Errorf("unsupported decision %q", contract.Decision)
	}
}

// publishError marks GitHub write failures as fatal so Claude does not retry them.
func publishError(err error) error {
	return fmt.Errorf("%w: publish resolution to GitHub: %v", claude.ErrAbort, err)
}

func (executor *Executor) HandlePrematureClosure(ctx context.Context) error {
	issue, err := executor.GitHub.GetIssue(ctx, executor.IssueNumber)
	if err != nil {
		return err
	}
	if issue.State != "closed" || hasLabel(issue.Labels, "vi:done") {
		return nil
	}
	if err := executor.GitHub.ReopenIssue(ctx, issue.Number); err != nil {
		return err
	}
	return executor.GitHub.AddIssueComment(ctx, issue.Number, "This VI was reopened because its documentation-resolution gate is not complete. Resolve the agent's outstanding questions or merge the docs PR, then apply the appropriate `vi:*` label.")
}

func (executor *Executor) ToolTraceMarkdown(calls []claude.ToolCall) string {
	var lines []string
	for index, call := range calls {
		lines = append(lines, fmt.Sprintf("%d. `%s`", index+1, call.Name))
	}
	return strings.Join(lines, "\n")
}

func (executor *Executor) inspectChanges(files []gh.PullRequestFile) []string {
	return changedConcepts(files, executor.Mapping)
}

func changedConcepts(files []gh.PullRequestFile, mapping documents.Mapping) []string {
	allConcepts := make(map[string]struct{})
	for _, feature := range mapping.Features {
		for _, concept := range feature.Schema {
			allConcepts[concept] = struct{}{}
		}
	}
	concepts := make([]string, 0, len(allConcepts))
	for concept := range allConcepts {
		concepts = append(concepts, concept)
	}
	sort.Strings(concepts)
	found := make(map[string]struct{})
	for _, file := range files {
		patch := file.Patch
		if strings.HasSuffix(file.Filename, "/config.yaml") && strings.Contains(file.Filename, "process_monitoring") {
			for _, concept := range concepts {
				if concept != "process_monitoring" && identifierIn(concept, patch) {
					found[concept] = struct{}{}
				}
			}
			continue
		}
		for _, line := range strings.Split(patch, "\n") {
			if !strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "-") {
				continue
			}
			if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
				continue
			}
			for _, concept := range concepts {
				if concept != "process_monitoring" && identifierIn(concept, line) {
					found[concept] = struct{}{}
				}
			}
		}
	}
	result := make([]string, 0, len(found))
	for concept := range found {
		result = append(result, concept)
	}
	sort.Strings(result)
	return result
}

func diffContext(files []gh.PullRequestFile, concepts []string) []gh.PullRequestFile {
	conceptSet := make(map[string]struct{}, len(concepts))
	for _, concept := range concepts {
		conceptSet[concept] = struct{}{}
	}
	result := make([]gh.PullRequestFile, 0, len(files))
	for _, file := range files {
		contextFile := gh.PullRequestFile{Filename: file.Filename, Status: file.Status}
		if len(concepts) > 0 {
			relevant := strings.HasPrefix(file.Filename, "product/process_monitoring/")
			if !relevant {
				for concept := range conceptSet {
					if identifierIn(concept, file.Patch) {
						relevant = true
						break
					}
				}
			}
			if relevant {
				contextFile.Patch = truncate(file.Patch, 4000)
			}
		}
		result = append(result, contextFile)
	}
	return result
}

func containsPatch(files []gh.PullRequestFile, filename string) bool {
	for _, file := range files {
		if file.Filename == filename && file.Patch != "" {
			return true
		}
	}
	return false
}

func identifierIn(identifier, text string) bool {
	return regexp.MustCompile(`(^|[^A-Za-z0-9_])` + regexp.QuoteMeta(identifier) + `([^A-Za-z0-9_]|$)`).MatchString(text)
}

func implementationPRNumber(body string) (int, error) {
	lines := strings.Split(body, "\n")
	for index, line := range lines {
		if !strings.Contains(strings.ToLower(line), "implementation pr") {
			continue
		}
		for next := index + 1; next < len(lines); next++ {
			value := strings.TrimSpace(lines[next])
			if value == "" {
				continue
			}
			match := issueNumberPattern.FindStringSubmatch(value)
			if len(match) > 1 {
				number, err := strconv.Atoi(match[1])
				if err == nil && number > 0 {
					return number, nil
				}
			}
			break
		}
	}
	return 0, fmt.Errorf("VI body must include a merged implementation PR number")
}

func issueWorkItem(issue gh.Issue) documents.WorkItem {
	return documents.WorkItem{
		Number:          issue.Number,
		Title:           issue.Title,
		URL:             issue.HTMLURL,
		UserGoal:        issueField(issue.Body, "User goal"),
		Audience:        issueField(issue.Body, "Target audience"),
		ExpectedOutcome: issueField(issue.Body, "Expected customer outcome"),
		Limitations:     nonemptyLines(issueField(issue.Body, "Limitations")),
	}
}

func issueField(body, field string) string {
	lines := strings.Split(body, "\n")
	active := false
	var value []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### ") {
			if active {
				break
			}
			active = strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(trimmed, "### ")), field)
			continue
		}
		if active {
			value = append(value, line)
		}
	}
	return strings.TrimSpace(strings.Join(value, "\n"))
}

func nonemptyLines(value string) []string {
	var result []string
	for _, line := range strings.Split(value, "\n") {
		if trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- ")); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (executor *Executor) addEvidence(evidence documents.Evidence) {
	executor.Evidence[evidence.ID] = evidence
}

func (executor *Executor) evidenceList() []documents.Evidence {
	ids := make([]string, 0, len(executor.Evidence))
	for id := range executor.Evidence {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]documents.Evidence, 0, len(ids))
	for _, id := range ids {
		result = append(result, executor.Evidence[id])
	}
	return result
}

func (executor *Executor) canonicalEvidence(proposed []documents.Evidence) []documents.Evidence {
	if len(proposed) == 0 {
		return nil
	}
	var result []documents.Evidence
	seen := make(map[string]struct{})
	for _, item := range proposed {
		authoritative, ok := executor.Evidence[item.ID]
		if !ok {
			return nil
		}
		if _, exists := seen[item.ID]; exists {
			continue
		}
		seen[item.ID] = struct{}{}
		result = append(result, authoritative)
	}
	return result
}

func (executor *Executor) contextMeasurement(selected map[string]struct{}) ContextMeasurement {
	measurement := ContextMeasurement{SelectedFiles: len(selected)}
	for path := range selected {
		if content, err := os.ReadFile(filepath.Join(executor.RepoRoot, path)); err == nil {
			measurement.SelectedChars += len(content)
		}
	}
	command := exec.Command("git", "ls-files", "-z")
	command.Dir = executor.RepoRoot
	if output, err := command.Output(); err == nil {
		for _, item := range strings.Split(string(output), "\x00") {
			if item == "" {
				continue
			}
			if info, err := os.Stat(filepath.Join(executor.RepoRoot, item)); err == nil && !info.IsDir() {
				measurement.RepositoryFiles++
				measurement.CandidateBytes += info.Size()
			}
		}
	}
	if measurement.CandidateBytes > 0 {
		selectedBytes := int64(measurement.SelectedChars)
		measurement.ReductionPct = 100 * (1 - float64(selectedBytes)/float64(measurement.CandidateBytes))
		if measurement.ReductionPct < 0 {
			measurement.ReductionPct = 0
		}
	}
	return measurement
}

func (executor *Executor) saveContract(contract documents.Contract) error {
	if err := os.MkdirAll(filepath.Join(executor.RepoRoot, "artifacts"), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		return fmt.Errorf("encode Doc Contract: %w", err)
	}
	path := filepath.Join(executor.RepoRoot, "artifacts", fmt.Sprintf("doc-contract-%d.json", executor.Issue.Number))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("save Doc Contract: %w", err)
	}
	return nil
}

func (executor *Executor) appendSummary(title string, contract documents.Contract, docsPR string) {
	path := os.Getenv("GITHUB_STEP_SUMMARY")
	if path == "" {
		return
	}
	var content strings.Builder
	fmt.Fprintf(&content, "## Documentation resolution: %s\n\n", title)
	fmt.Fprintf(&content, "- VI: [#%d](%s)\n", executor.Issue.Number, executor.Issue.HTMLURL)
	fmt.Fprintf(&content, "- Implementation PR: [#%d](%s)\n", executor.Pull.Number, executor.Pull.HTMLURL)
	fmt.Fprintf(&content, "- Provider: Anthropic Claude (`%s`)\n", executor.Model)
	if executor.Gateway != "" {
		fmt.Fprintf(&content, "- Gateway: `%s`\n", executor.Gateway)
	}
	fmt.Fprintf(&content, "- Decision: `%s`\n", contract.Decision)
	fmt.Fprintf(&content, "- Context: %d selected files, approximately %d characters from %d tracked files (%.1f%% reduction)\n", executor.measurement.SelectedFiles, executor.measurement.SelectedChars, executor.measurement.RepositoryFiles, executor.measurement.ReductionPct)
	if docsPR != "" {
		fmt.Fprintf(&content, "- Docs PR: %s\n", docsPR)
	}
	content.WriteString("\n### Affected documentation\n\n")
	if len(contract.AffectedDocuments) == 0 {
		content.WriteString("No docs proposals.\n")
	}
	for _, proposal := range contract.AffectedDocuments {
		fmt.Fprintf(&content, "- `%s` (%s / %s): %s\n", proposal.Path, proposal.Audience, proposal.Type, proposal.Reason)
	}
	if len(contract.Conflicts) > 0 {
		content.WriteString("\n### Conflicts\n\n")
		for _, conflict := range contract.Conflicts {
			fmt.Fprintf(&content, "- %s (blocking: %t)\n", conflict.Description, conflict.BlocksResolution)
		}
	}
	if len(contract.MissingInformation) > 0 {
		content.WriteString("\n### Missing information\n\n")
		for _, item := range contract.MissingInformation {
			fmt.Fprintf(&content, "- %s\n", item)
		}
	}
	content.WriteString("\nHuman review is required; documentation is never auto-merged.\n")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.WriteString(content.String())
}

func resolutionComment(contract documents.Contract) string {
	var comment strings.Builder
	comment.WriteString("## Documentation resolution needs clarification\n\n")
	for _, conflict := range contract.Conflicts {
		fmt.Fprintf(&comment, "- **Conflict:** %s\n", conflict.Description)
	}
	for _, missing := range contract.MissingInformation {
		fmt.Fprintf(&comment, "- **Missing information:** %s\n", missing)
	}
	comment.WriteString("\nThe VI remains open for human resolution. Evidence is available in the workflow artifact.")
	return comment.String()
}

func docsPRBody(contract documents.Contract, implementation gh.PullRequest) string {
	var body strings.Builder
	fmt.Fprintf(&body, "## Trigger\n\nThis proposal resolves VI #%d using merged implementation PR #%d.\n\n", contract.WorkItem.Number, implementation.Number)
	body.WriteString("## Affected documents\n\n")
	for _, document := range contract.AffectedDocuments {
		fmt.Fprintf(&body, "- `%s` (%s / %s): %s\n", document.Path, document.Audience, document.Type, document.Reason)
	}
	body.WriteString("\n## Evidence and unresolved items\n\n")
	for _, claim := range contract.Claims {
		fmt.Fprintf(&body, "- [%s] %s (evidence: %s)\n", claim.Audience, claim.Text, strings.Join(claim.EvidenceIDs, ", "))
	}
	if len(contract.Conflicts) == 0 && len(contract.MissingInformation) == 0 {
		body.WriteString("No unresolved conflicts or missing information were identified.\n")
	}
	for _, conflict := range contract.Conflicts {
		fmt.Fprintf(&body, "- **Conflict:** %s (evidence: %s)\n", conflict.Description, strings.Join(conflict.EvidenceIDs, ", "))
	}
	for _, missing := range contract.MissingInformation {
		fmt.Fprintf(&body, "- **Missing:** %s\n", missing)
	}
	body.WriteString("\nHuman review is required. This PR is not auto-merged.\n")
	return body.String()
}

func requireAllSelectedDocuments(proposals []documents.DocumentProposal, targets []documents.DocumentTarget) error {
	targetPaths := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		targetPaths[target.Path] = struct{}{}
	}
	proposed := make(map[string]struct{}, len(proposals))
	for _, proposal := range proposals {
		if _, ok := targetPaths[proposal.Path]; !ok {
			return fmt.Errorf("proposal includes document %s that was not selected for this change", proposal.Path)
		}
		proposed[proposal.Path] = struct{}{}
	}
	for _, target := range targets {
		if _, ok := proposed[target.Path]; !ok {
			return fmt.Errorf("proposal is missing selected %s document %s", target.Audience, target.Path)
		}
	}
	return nil
}

func validateSettingNames(settings []documents.SettingChange, mapping documents.Mapping) error {
	allowed := make(map[string]struct{})
	for _, feature := range mapping.Features {
		for _, concept := range feature.Schema {
			allowed[concept] = struct{}{}
		}
	}
	for _, setting := range settings {
		if _, ok := allowed[setting.Name]; !ok {
			return fmt.Errorf("contract references setting %q that is not in the docs mapping", setting.Name)
		}
	}
	return nil
}

func audiences(targets []documents.DocumentTarget) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, target := range targets {
		if _, ok := seen[target.Audience]; ok {
			continue
		}
		seen[target.Audience] = struct{}{}
		result = append(result, target.Audience)
	}
	sort.Strings(result)
	return result
}

func containsAudience(targets []documents.DocumentTarget, audience string) bool {
	for _, target := range targets {
		if target.Audience == audience {
			return true
		}
	}
	return false
}

func hasLabel(labels []gh.Label, name string) bool {
	for _, label := range labels {
		if label.Name == name {
			return true
		}
	}
	return false
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "\n[truncated]"
}

func safeID(value string) string {
	value = strings.ToUpper(value)
	return strings.NewReplacer("/", "-", ".", "-", "_", "-").Replace(value)
}

func (executor *Executor) readLocal(path string) (string, error) {
	clean := filepath.Clean(path)
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path must stay inside repository")
	}
	data, err := os.ReadFile(filepath.Join(executor.RepoRoot, clean))
	return string(data), err
}
