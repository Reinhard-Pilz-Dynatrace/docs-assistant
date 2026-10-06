package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultEndpoint = "https://api.anthropic.com/v1/messages"

type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type ToolExecutor interface {
	Definitions() []ToolDefinition
	Execute(context.Context, string, json.RawMessage) (any, error)
}

type Client struct {
	APIKey       string // sent as x-api-key
	AuthToken    string // sent as Authorization: Bearer (gateway credential)
	BaseURL      string // optional gateway base URL, e.g. https://gateway.example.com
	Model        string
	Endpoint     string
	HTTPClient   *http.Client
	MaxTokens    int
	MaxToolTurns int
}

// ErrAbort marks a tool error that Claude cannot fix by retrying (for example a
// failed GitHub write). Run stops and returns it instead of feeding it back.
var ErrAbort = errors.New("agent run aborted")

type ToolCall struct {
	Name   string          `json:"name"`
	Input  json.RawMessage `json:"input"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

type RunResult struct {
	Text      string     `json:"text"`
	Model     string     `json:"model"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

type apiRequest struct {
	Model      string           `json:"model"`
	MaxTokens  int              `json:"max_tokens"`
	System     string           `json:"system"`
	Tools      []ToolDefinition `json:"tools"`
	Messages   []message        `json:"messages"`
	ToolChoice *toolChoice      `json:"tool_choice,omitempty"`
}

type toolChoice struct {
	Type string `json:"type"`
}

type message struct {
	Role    string        `json:"role"`
	Content []contentPart `json:"content"`
}

type contentPart struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
}

type apiResponse struct {
	Model      string        `json:"model"`
	StopReason string        `json:"stop_reason"`
	Content    []contentPart `json:"content"`
	Error      *apiError     `json:"error,omitempty"`
}

type apiError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (client Client) Run(ctx context.Context, systemPrompt, userPrompt string, executor ToolExecutor) (RunResult, error) {
	if client.APIKey == "" && client.AuthToken == "" {
		return RunResult{}, fmt.Errorf("ANTHROPIC_API_KEY or ANTHROPIC_AUTH_TOKEN is required for live agent runs")
	}
	if client.Model == "" {
		return RunResult{}, fmt.Errorf("ANTHROPIC_MODEL is required for live agent runs")
	}
	if executor == nil || len(executor.Definitions()) == 0 {
		return RunResult{}, fmt.Errorf("at least one agent tool is required")
	}

	endpoint := client.Endpoint
	if endpoint == "" {
		endpoint = defaultEndpoint
		if client.BaseURL != "" {
			endpoint = strings.TrimRight(client.BaseURL, "/") + "/v1/messages"
		}
	}
	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 90 * time.Second}
	}
	maxTokens := client.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 8192
	}
	maxTurns := client.MaxToolTurns
	if maxTurns <= 0 {
		maxTurns = 8
	}

	result := RunResult{Model: client.Model}
	messages := []message{{Role: "user", Content: []contentPart{{Type: "text", Text: userPrompt}}}}
	for turn := 0; turn < maxTurns; turn++ {
		requestBody := apiRequest{
			Model:      client.Model,
			MaxTokens:  maxTokens,
			System:     systemPrompt,
			Tools:      executor.Definitions(),
			Messages:   messages,
			ToolChoice: &toolChoice{Type: "auto"},
		}
		encoded, err := json.Marshal(requestBody)
		if err != nil {
			return result, fmt.Errorf("encode Claude request: %w", err)
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
		if err != nil {
			return result, fmt.Errorf("create Claude request: %w", err)
		}
		if client.AuthToken != "" {
			request.Header.Set("Authorization", "Bearer "+client.AuthToken)
		} else {
			request.Header.Set("x-api-key", client.APIKey)
		}
		request.Header.Set("anthropic-version", "2023-06-01")
		request.Header.Set("content-type", "application/json")

		response, err := httpClient.Do(request)
		if err != nil {
			return result, fmt.Errorf("call Claude Messages API: %w", err)
		}
		body, readErr := io.ReadAll(io.LimitReader(response.Body, 4<<20))
		closeErr := response.Body.Close()
		if readErr != nil {
			return result, fmt.Errorf("read Claude response: %w", readErr)
		}
		if closeErr != nil {
			return result, fmt.Errorf("close Claude response: %w", closeErr)
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return result, fmt.Errorf("Claude API returned %s: %s", response.Status, strings.TrimSpace(string(body)))
		}

		var reply apiResponse
		if err := json.Unmarshal(body, &reply); err != nil {
			return result, fmt.Errorf("decode Claude response: %w", err)
		}
		if reply.Error != nil {
			return result, fmt.Errorf("Claude API error %s: %s", reply.Error.Type, reply.Error.Message)
		}
		result.Model = reply.Model
		for _, part := range reply.Content {
			if part.Type == "text" {
				result.Text += part.Text
			}
		}
		messages = append(messages, message{Role: "assistant", Content: reply.Content})
		if reply.StopReason != "tool_use" {
			if result.Text == "" {
				return result, fmt.Errorf("Claude ended without a final text response")
			}
			return result, nil
		}

		var toolResults []contentPart
		for _, part := range reply.Content {
			if part.Type != "tool_use" {
				continue
			}
			toolValue, executeErr := executor.Execute(ctx, part.Name, part.Input)
			if executeErr != nil {
				toolValue = map[string]string{"error": executeErr.Error()}
			}
			toolContent, marshalErr := json.Marshal(toolValue)
			if marshalErr != nil {
				return result, fmt.Errorf("encode result for tool %q: %w", part.Name, marshalErr)
			}
			toolCall := ToolCall{Name: part.Name, Input: part.Input, Result: toolContent}
			if executeErr != nil {
				toolCall.Error = executeErr.Error()
			}
			result.ToolCalls = append(result.ToolCalls, toolCall)
			if errors.Is(executeErr, ErrAbort) {
				return result, executeErr
			}
			toolResults = append(toolResults, contentPart{
				Type:      "tool_result",
				ToolUseID: part.ID,
				Content:   string(toolContent),
				IsError:   executeErr != nil,
			})
		}
		if len(toolResults) == 0 {
			return result, fmt.Errorf("Claude requested tool use without tool calls")
		}
		messages = append(messages, message{Role: "user", Content: toolResults})
	}
	return result, fmt.Errorf("Claude exceeded the maximum of %d tool turns", maxTurns)
}
