package documents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateContractRejectsUnknownEvidenceAndMissingTemplateSections(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs/templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs/templates/customer.md"), []byte("## Overview\n## Limitations\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := DocumentTarget{Path: "docs/customer.md", Audience: "customer", Type: "guide", Template: "docs/templates/customer.md"}
	mapping := Mapping{Features: map[string]FeatureMapping{"feature": {Schema: []string{"setting"}, Docs: []DocumentTarget{target}}}}
	base := Contract{
		Version:  "1",
		Decision: DecisionProposeDocs,
		Evidence: []Evidence{{ID: "E1", Type: "CODE", Source: "product/config.go", Detail: "setting=true"}},
		Claims:   []Claim{{Text: "The setting is enabled.", Audience: "customer", EvidenceIDs: []string{"E1"}}},
		AffectedDocuments: []DocumentProposal{{
			DocumentTarget: target,
			Markdown:       "# Feature\n\n## Overview\nSupported.\n",
		}},
	}

	if err := ValidateContract(base, mapping, root); err == nil || !strings.Contains(err.Error(), "Limitations") {
		t.Fatalf("ValidateContract() error = %v, want missing Limitations heading", err)
	}

	base.AffectedDocuments[0].Markdown += "\n## Limitations\nNone known.\n"
	base.Claims[0].EvidenceIDs = []string{"MISSING"}
	if err := ValidateContract(base, mapping, root); err == nil || !strings.Contains(err.Error(), "unknown evidence") {
		t.Fatalf("ValidateContract() error = %v, want unknown evidence error", err)
	}
}

func TestValidateContractRejectsUnmappedDocument(t *testing.T) {
	mapping := Mapping{Features: map[string]FeatureMapping{"feature": {Schema: []string{"setting"}, Docs: []DocumentTarget{{Path: "docs/customer.md", Audience: "customer", Type: "guide", Template: "docs/templates/customer.md"}}}}}
	contract := Contract{Version: "1", Decision: DecisionProposeDocs, AffectedDocuments: []DocumentProposal{{DocumentTarget: DocumentTarget{Path: "docs/other.md", Audience: "customer", Type: "guide", Template: "docs/templates/customer.md"}, Markdown: "# Other"}}}
	if err := ValidateContract(contract, mapping, t.TempDir()); err == nil || !strings.Contains(err.Error(), "not an allowed mapped target") {
		t.Fatalf("ValidateContract() error = %v, want unmapped target error", err)
	}
}
