package workflow

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/documents"
	gh "github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/github"
)

func TestChangedConceptsSelectsSettingsButNotUnrelatedRefactor(t *testing.T) {
	mapping := documents.Mapping{Features: map[string]documents.FeatureMapping{
		"process_monitoring": {Schema: []string{"process_monitoring", "include_container_processes", "scan_interval_seconds"}, Source: []string{"product/process_monitoring/config.yaml", "product/process_monitoring/process_monitor.go"}},
	}}
	files := []gh.PullRequestFile{
		{Filename: "product/process_monitoring/config.yaml", Patch: "@@ -1,2 +1,3 @@ process_monitoring:\n   enabled: true\n+  include_container_processes: true\n"},
		{Filename: "product/process_monitoring/process_monitor.go", Patch: "@@ -2 +2 @@\n-func oldName() {}\n+func normalizedName() {}\n"},
	}
	got := changedConcepts(files, mapping)
	if len(got) != 1 || got[0] != "include_container_processes" {
		t.Fatalf("changedConcepts() = %#v, want only include_container_processes", got)
	}
}

func TestDiffContextBoundsMappedPatchesAndOmitsUnrelatedPatches(t *testing.T) {
	files := []gh.PullRequestFile{
		{Filename: "product/process_monitoring/config.yaml", Status: "modified", Patch: strings.Repeat("+ include_container_processes\n", 300)},
		{Filename: "docs/README.md", Status: "modified", Patch: "+ unrelated content\n"},
	}
	got := diffContext(files, []string{"include_container_processes"}, testMapping())
	if len(got) != 2 || len(got[0].Patch) > 4020 || got[1].Patch != "" {
		t.Fatalf("diffContext() = %#v", got)
	}
	noImpact := diffContext(files, nil, testMapping())
	if noImpact[0].Patch != "" || noImpact[1].Patch != "" {
		t.Fatalf("no-impact context included patches: %#v", noImpact)
	}
}

func TestImplementationPRNumberReadsIssueFormField(t *testing.T) {
	body := "### User goal\n\nInclude container processes.\n\n### Merged implementation PR\n\n#42\n"
	got, err := implementationPRNumber(body)
	if err != nil {
		t.Fatalf("implementationPRNumber() error = %v", err)
	}
	if got != 42 {
		t.Fatalf("implementationPRNumber() = %d, want 42", got)
	}
}

func TestReadContextForNoImpactRetainsEvidence(t *testing.T) {
	executor := NewExecutor(nil, documents.Mapping{}, 7, t.TempDir(), "test-model")
	executor.Issue = gh.Issue{Number: 7}
	executor.Pull = gh.PullRequest{MergeCommitSHA: "merge-sha"}
	executor.addEvidence(documents.Evidence{ID: "VI-7", Type: "VI", Source: "issue", Detail: "intent"})

	bundle, err := executor.readDocumentationContext(context.Background())
	if err != nil {
		t.Fatalf("readDocumentationContext() error = %v", err)
	}
	if len(bundle.SelectedDocuments) != 0 || len(bundle.Evidence) != 1 || bundle.Evidence[0].ID != "VI-7" {
		t.Fatalf("no-impact context = %#v", bundle)
	}
}

func TestReadDocumentationContextLoadsOnlyMappedAudienceFiles(t *testing.T) {
	files := map[string]string{
		"product/process_monitoring/config.yaml":             "process_monitoring:\n  scan_interval_seconds: 30\n",
		"product/process_monitoring/process_monitor.go":      "package processmonitoring\n",
		"product/process_monitoring/process_monitor_test.go": "package processmonitoring\n",
		"docs/customer/process-monitoring.md":                "# Customer page\n",
		"docs/templates/customer.md":                         "## Overview\n",
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		path := strings.TrimPrefix(request.URL.Path, "/repos/acme/demo/contents/")
		content, ok := files[path]
		if !ok {
			http.NotFound(writer, request)
			return
		}
		encoded := base64.StdEncoding.EncodeToString([]byte(content))
		writer.Header().Set("content-type", "application/json")
		_, _ = fmt.Fprintf(writer, `{"path":%q,"sha":"sha-%s","encoding":"base64","content":%q}`, path, path, encoded)
	}))
	defer server.Close()

	target := documents.DocumentTarget{Path: "docs/customer/process-monitoring.md", Audience: "customer", Type: "product-guide", Template: "docs/templates/customer.md"}
	mapping := documents.Mapping{Features: map[string]documents.FeatureMapping{
		"process_monitoring": {Schema: []string{"scan_interval_seconds"}, Source: []string{"product/process_monitoring/config.yaml", "product/process_monitoring/process_monitor.go", "product/process_monitoring/process_monitor_test.go"}, Docs: []documents.DocumentTarget{target}},
	}}
	client := &gh.Client{Owner: "acme", Repository: "demo", Token: "test", APIBase: server.URL, HTTPClient: server.Client()}
	executor := NewExecutor(client, mapping, 7, t.TempDir(), "test-model")
	executor.Issue = gh.Issue{Number: 7}
	executor.Pull = gh.PullRequest{MergeCommitSHA: "merge-sha"}
	executor.Changed = []string{"scan_interval_seconds"}
	executor.addEvidence(documents.Evidence{ID: "VI-7", Type: "VI", Source: "issue", Detail: "intent"})

	bundle, err := executor.readDocumentationContext(context.Background())
	if err != nil {
		t.Fatalf("readDocumentationContext() error = %v", err)
	}
	if len(bundle.SelectedDocuments) != 1 || bundle.SelectedDocuments[0].Audience != "customer" {
		t.Fatalf("selected documents = %#v", bundle.SelectedDocuments)
	}
	if len(bundle.ExistingDocs) != 1 || len(bundle.Templates) != 1 || len(bundle.SourceFiles) != 3 {
		t.Fatalf("context bundle counts = docs %d, templates %d, source %d", len(bundle.ExistingDocs), len(bundle.Templates), len(bundle.SourceFiles))
	}
	if len(bundle.Evidence) != 6 {
		t.Fatalf("evidence count = %d, want 6", len(bundle.Evidence))
	}
}

func TestRequireAllSelectedDocumentsRejectsUnselectedExtra(t *testing.T) {
	target := documents.DocumentTarget{Path: "docs/customer.md"}
	extra := documents.DocumentProposal{DocumentTarget: documents.DocumentTarget{Path: "docs/internal.md"}}
	if err := requireAllSelectedDocuments([]documents.DocumentProposal{extra}, []documents.DocumentTarget{target}); err == nil {
		t.Fatal("requireAllSelectedDocuments() accepted an unselected document")
	}
}

func TestValidateSettingNamesRejectsUnmappedSetting(t *testing.T) {
	mapping := documents.Mapping{Features: map[string]documents.FeatureMapping{
		"process_monitoring": {Schema: []string{"scan_interval_seconds"}},
	}}
	err := validateSettingNames([]documents.SettingChange{{Name: "invented_setting", New: "true"}}, mapping)
	if err == nil || !strings.Contains(err.Error(), "not in the docs mapping") {
		t.Fatalf("validateSettingNames() error = %v, want unmapped setting error", err)
	}
}

func TestSubmitResolutionRejectsNoImpactWhenDocsWereSelected(t *testing.T) {
	target := documents.DocumentTarget{Path: "docs/customer.md", Audience: "customer", Type: "guide", Template: "docs/templates/customer.md"}
	executor := NewExecutor(nil, documents.Mapping{Features: map[string]documents.FeatureMapping{"feature": {Schema: []string{"setting"}, Docs: []documents.DocumentTarget{target}}}}, 7, t.TempDir(), "test-model")
	executor.ContextRead = true
	executor.Targets = []documents.DocumentTarget{target}
	evidence := documents.Evidence{ID: "VI-7", Type: "VI", Source: "issue", Detail: "intent"}
	executor.addEvidence(evidence)
	input, err := json.Marshal(map[string]any{"contract": documents.Contract{
		Version: "1", Decision: documents.DecisionNoDocsImpact, Evidence: []documents.Evidence{evidence},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := executor.submitResolution(context.Background(), input); err == nil || !strings.Contains(err.Error(), "mapping selected targets") {
		t.Fatalf("submitResolution() error = %v, want false no-impact rejection", err)
	}
}

func sampleContract() documents.Contract {
	return documents.Contract{
		Provider: documents.Provider{Name: "Anthropic Claude", Model: "test-model"},
		WorkItem: documents.WorkItem{Number: 4, Title: "VI: Faster | scans", UserGoal: "goal", ExpectedOutcome: "60 second default"},
		Change:   documents.Change{Settings: []documents.SettingChange{{Name: "scan_interval_seconds", Old: "30", New: "60"}}},
		Evidence: []documents.Evidence{
			{ID: "VI-4", Type: "VI", Source: "https://example/4", Detail: "d"},
			{ID: "SRC-PRODUCT-CONFIG-YAML", Type: "SCHEMA", Source: "product/process_monitoring/config.yaml@abc", Detail: "d"},
		},
		Claims:             []documents.Claim{{Text: "Default is 60 seconds.", Audience: "customer", EvidenceIDs: []string{"SRC-PRODUCT-CONFIG-YAML", "VI-4"}}},
		AffectedDocuments:  []documents.DocumentProposal{{DocumentTarget: documents.DocumentTarget{Path: "docs/a.md", Audience: "customer", Type: "product-guide", Reason: "settings-schema mapping"}}},
		Conflicts:          []documents.Conflict{{Description: "old docs say 30", EvidenceIDs: []string{"SRC-PRODUCT-CONFIG-YAML"}}},
		MissingInformation: []string{"Troubleshooting is not established."},
	}
}

func TestDocsPRBodyIsCompactAndReadable(t *testing.T) {
	measurement := ContextMeasurement{RepositoryFiles: 50, SelectedFiles: 7, ReductionPct: 86}
	body := docsPRBody(sampleContract(), gh.PullRequest{Number: 4}, measurement, "gateway.example")
	for _, want := range []string{
		"## Documentation update: VI #4 — Faster | scans",
		"Resolved by Claude (`test-model`) via `gateway.example`",
		"| `docs/a.md` | customer | settings-schema mapping |",
		"| `scan_interval_seconds` | 30 | 60 |",
		"- ? Troubleshooting is not established. `[MISSING]`",
		"7 of 50 repository files read",
		"<summary>Evidence (1 claims)</summary>\n\n",
		"- Default is 60 seconds. — `SCHEMA config.yaml`, `VI #4`",
		"<summary>Conflicts and notes</summary>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body is missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "SRC-PRODUCT") {
		t.Errorf("body leaks raw evidence IDs:\n%s", body)
	}
}

func TestBlockingConflictBecomesWarningAndClipIsBounded(t *testing.T) {
	contract := sampleContract()
	contract.Conflicts = []documents.Conflict{{Description: "VI says 50\nMB", BlocksResolution: true}}
	if body := docsPRBody(contract, gh.PullRequest{Number: 4}, ContextMeasurement{}, ""); !strings.Contains(body, "> [!WARNING]\n> **Source conflict:** VI says 50 MB") {
		t.Fatalf("blocking conflict not rendered as warning:\n%s", body)
	}
	if got := clip(strings.Repeat("x", maxReportChars+10)); len(got) > maxReportChars+200 {
		t.Fatalf("clip() length = %d", len(got))
	}
}

func TestResolutionCommentShowsConflictTableAndQuestion(t *testing.T) {
	contract := sampleContract()
	comment := resolutionComment(contract)
	for _, want := range []string{"## Needs clarification", "| Source conflict | Evidence |", "| old docs say 30 | `SCHEMA config.yaml` |", "### Question for a human", "re-apply `vi:ready-for-docs`"} {
		if !strings.Contains(comment, want) {
			t.Errorf("comment is missing %q:\n%s", want, comment)
		}
	}
}

func testMapping() documents.Mapping {
	return documents.Mapping{Features: map[string]documents.FeatureMapping{
		"process_monitoring": {Schema: []string{"process_monitoring", "include_container_processes"}, Source: []string{"product/process_monitoring/config.yaml"}},
	}}
}

func TestChangedConceptsSeparatesFeatures(t *testing.T) {
	mapping := documents.Mapping{Features: map[string]documents.FeatureMapping{
		"log_collection":     {Schema: []string{"log_collection", "max_file_size_mb"}, Source: []string{"product/log_collection/config.yaml"}},
		"network_monitoring": {Schema: []string{"network_monitoring", "sample_rate_percent"}, Source: []string{"product/network_monitoring/config.yaml"}},
	}}
	files := []gh.PullRequestFile{{Filename: "product/log_collection/config.yaml", Patch: "@@ -1,3 +1,3 @@ log_collection:\n-  max_file_size_mb: 100\n+  max_file_size_mb: 50\n"}}
	got := changedConcepts(files, mapping)
	if len(got) != 1 || got[0] != "max_file_size_mb" {
		t.Fatalf("changedConcepts() = %#v", got)
	}
	if features := mapping.FeaturesForConcepts(got); len(features) != 1 || features[0] != "log_collection" {
		t.Fatalf("FeaturesForConcepts() = %#v", features)
	}
}
