package claude

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testExecutor struct {
	calls []string
	err   error
}

func (executor *testExecutor) Definitions() []ToolDefinition {
	return []ToolDefinition{{Name: "inspect_vi", Description: "Read VI evidence", InputSchema: json.RawMessage(`{"type":"object","properties":{}}`)}}
}

func (executor *testExecutor) Execute(_ context.Context, name string, _ json.RawMessage) (any, error) {
	executor.calls = append(executor.calls, name)
	if executor.err != nil {
		return nil, executor.err
	}
	return map[string]string{"evidence": "VI body"}, nil
}

func TestRunExecutesToolAndContinuesConversation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("x-api-key") != "test-key" {
			t.Errorf("x-api-key = %q, want test key", request.Header.Get("x-api-key"))
		}
		if request.Header.Get("anthropic-version") == "" {
			t.Error("anthropic-version header is missing")
		}
		var body apiRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		writer.Header().Set("content-type", "application/json")
		if len(body.Messages) == 1 {
			_, _ = writer.Write([]byte(`{"model":"test-model","stop_reason":"tool_use","content":[{"type":"tool_use","id":"call-1","name":"inspect_vi","input":{}}]}`))
			return
		}
		if len(body.Messages) != 3 || body.Messages[2].Content[0].Type != "tool_result" {
			t.Errorf("tool result was not returned in next request: %#v", body.Messages)
		}
		_, _ = writer.Write([]byte(`{"model":"test-model","stop_reason":"end_turn","content":[{"type":"text","text":"VI inspected."}]}`))
	}))
	defer server.Close()

	executor := &testExecutor{}
	client := Client{APIKey: "test-key", Model: "test-model", Endpoint: server.URL, HTTPClient: server.Client()}
	result, err := client.Run(context.Background(), "system", "inspect the VI", executor)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Text != "VI inspected." || result.Model != "test-model" {
		t.Fatalf("Run() result = %#v", result)
	}
	if len(executor.calls) != 1 || executor.calls[0] != "inspect_vi" {
		t.Fatalf("tool calls = %#v", executor.calls)
	}
	if len(result.ToolCalls) != 1 || string(result.ToolCalls[0].Result) != `{"evidence":"VI body"}` {
		t.Fatalf("tool trace = %#v", result.ToolCalls)
	}
}

func TestRunRequiresLiveCredentials(t *testing.T) {
	client := Client{}
	_, err := client.Run(context.Background(), "system", "user", &testExecutor{})
	if err == nil || !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Fatalf("Run() error = %v, want missing API key error", err)
	}
}

func TestRunReturnsToolErrorsToClaude(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body apiRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		writer.Header().Set("content-type", "application/json")
		if len(body.Messages) == 1 {
			_, _ = writer.Write([]byte(`{"model":"test-model","stop_reason":"tool_use","content":[{"type":"tool_use","id":"call-1","name":"inspect_vi","input":{}}]}`))
			return
		}
		toolResult := body.Messages[2].Content[0]
		if !toolResult.IsError || !strings.Contains(toolResult.Content, "permission denied") {
			t.Errorf("tool error result = %#v", toolResult)
		}
		_, _ = writer.Write([]byte(`{"model":"test-model","stop_reason":"end_turn","content":[{"type":"text","text":"I will report the tool failure."}]}`))
	}))
	defer server.Close()

	executor := &testExecutor{err: errors.New("permission denied")}
	client := Client{APIKey: "test-key", Model: "test-model", Endpoint: server.URL, HTTPClient: server.Client()}
	if _, err := client.Run(context.Background(), "system", "inspect", executor); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}
