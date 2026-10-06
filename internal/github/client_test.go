package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetWorkflowLabelPreservesOtherLabelsAndReplacesState(t *testing.T) {
	var updated []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/acme/demo/labels/vi:ready-for-docs":
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(`{"name":"vi:ready-for-docs"}`))
		case request.Method == http.MethodGet && request.URL.Path == "/repos/acme/demo/issues/7":
			_, _ = writer.Write([]byte(`{"number":7,"labels":[{"name":"team:agent"},{"name":"vi:in-progress"}]}`))
		case request.Method == http.MethodPut && request.URL.Path == "/repos/acme/demo/issues/7/labels":
			var body struct {
				Labels []string `json:"labels"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Errorf("decode labels: %v", err)
			}
			updated = body.Labels
			_, _ = writer.Write([]byte(`[]`))
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL)
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := &Client{Owner: "acme", Repository: "demo", Token: "test", APIBase: server.URL, HTTPClient: server.Client()}
	if err := client.SetWorkflowLabel(context.Background(), 7, "vi:ready-for-docs"); err != nil {
		t.Fatalf("SetWorkflowLabel() error = %v", err)
	}
	if len(updated) != 2 || updated[0] != "team:agent" || updated[1] != "vi:ready-for-docs" {
		t.Fatalf("updated labels = %#v", updated)
	}
}

func TestSetWorkflowLabelRejectsUnknownState(t *testing.T) {
	client := &Client{Owner: "acme", Repository: "demo"}
	if err := client.SetWorkflowLabel(context.Background(), 7, "ready"); err == nil {
		t.Fatal("SetWorkflowLabel() accepted an unknown label")
	}
}

func TestCreateDocumentationPRCreatesBranchUpdatesMappedFileAndOpensPR(t *testing.T) {
	updatedContent := ""
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/acme/demo":
			_, _ = writer.Write([]byte(`{"default_branch":"main"}`))
		case request.Method == http.MethodGet && request.URL.Path == "/repos/acme/demo/pulls":
			_, _ = writer.Write([]byte(`[]`))
		case request.Method == http.MethodGet && request.URL.Path == "/repos/acme/demo/git/ref/heads/docs/vi-9":
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte(`{"message":"Not Found"}`))
		case request.Method == http.MethodGet && request.URL.Path == "/repos/acme/demo/branches/main":
			_, _ = writer.Write([]byte(`{"commit":{"sha":"base-sha"}}`))
		case request.Method == http.MethodPost && request.URL.Path == "/repos/acme/demo/git/refs":
			var body map[string]string
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body["ref"] != "refs/heads/docs/vi-9" || body["sha"] != "base-sha" {
				t.Errorf("create ref payload = %#v", body)
			}
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{}`))
		case request.Method == http.MethodGet && request.URL.Path == "/repos/acme/demo/contents/docs/customer.md":
			if request.URL.Query().Get("ref") != "docs/vi-9" {
				t.Errorf("content ref = %q", request.URL.Query().Get("ref"))
			}
			content := base64.StdEncoding.EncodeToString([]byte("old docs"))
			_, _ = writer.Write([]byte(`{"sha":"file-sha","encoding":"base64","content":"` + content + `"}`))
		case request.Method == http.MethodPut && request.URL.Path == "/repos/acme/demo/contents/docs/customer.md":
			var body map[string]string
			_ = json.NewDecoder(request.Body).Decode(&body)
			decoded, err := base64.StdEncoding.DecodeString(body["content"])
			if err != nil {
				t.Errorf("decode update content: %v", err)
			}
			updatedContent = string(decoded)
			if body["sha"] != "file-sha" || body["branch"] != "docs/vi-9" {
				t.Errorf("update file payload = %#v", body)
			}
			_, _ = writer.Write([]byte(`{}`))
		case request.Method == http.MethodPost && request.URL.Path == "/repos/acme/demo/pulls":
			var body map[string]string
			_ = json.NewDecoder(request.Body).Decode(&body)
			if !strings.Contains(body["body"], "human review") || body["base"] != "main" {
				t.Errorf("create pull request payload = %#v", body)
			}
			_, _ = writer.Write([]byte(`{"number":19,"html_url":"https://github.com/acme/demo/pull/19"}`))
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL)
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := &Client{Owner: "acme", Repository: "demo", Token: "test", APIBase: server.URL, HTTPClient: server.Client()}
	pull, err := client.CreateDocumentationPR(context.Background(), 9, "docs update", "Requires human review", map[string]string{"docs/customer.md": "new docs"})
	if err != nil {
		t.Fatalf("CreateDocumentationPR() error = %v", err)
	}
	if pull.Number != 19 || pull.HTMLURL != "https://github.com/acme/demo/pull/19" || updatedContent != "new docs" {
		t.Fatalf("pull = %#v, updated content = %q", pull, updatedContent)
	}
}
