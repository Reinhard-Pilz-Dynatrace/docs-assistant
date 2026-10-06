package documents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var evidenceTypes = map[string]struct{}{
	"CODE":          {},
	"SCHEMA":        {},
	"TEST":          {},
	"VI":            {},
	"EXISTING_DOCS": {},
	"INFERRED":      {},
}

func ValidateContract(contract Contract, mapping Mapping, repoRoot string) error {
	if contract.Version == "" {
		return fmt.Errorf("contract version is required")
	}
	if contract.Decision != DecisionProposeDocs && contract.Decision != DecisionNeedsClarification && contract.Decision != DecisionNoDocsImpact {
		return fmt.Errorf("unsupported decision %q", contract.Decision)
	}

	evidenceByID := make(map[string]Evidence, len(contract.Evidence))
	for _, evidence := range contract.Evidence {
		if evidence.ID == "" || evidence.Source == "" || evidence.Detail == "" {
			return fmt.Errorf("evidence entries require id, source, and detail")
		}
		if _, ok := evidenceTypes[evidence.Type]; !ok {
			return fmt.Errorf("evidence %q has unsupported type %q", evidence.ID, evidence.Type)
		}
		if _, exists := evidenceByID[evidence.ID]; exists {
			return fmt.Errorf("duplicate evidence id %q", evidence.ID)
		}
		evidenceByID[evidence.ID] = evidence
	}

	for _, claim := range contract.Claims {
		if claim.Text == "" || (claim.Audience != "customer" && claim.Audience != "internal-developer") {
			return fmt.Errorf("claims require text and a supported audience")
		}
		if err := validateEvidenceReferences(claim.EvidenceIDs, evidenceByID, "claim"); err != nil {
			return err
		}
	}
	for _, conflict := range contract.Conflicts {
		if conflict.Description == "" {
			return fmt.Errorf("conflicts require a description")
		}
		if err := validateEvidenceReferences(conflict.EvidenceIDs, evidenceByID, "conflict"); err != nil {
			return err
		}
	}

	if contract.Decision == DecisionNeedsClarification || contract.Decision == DecisionNoDocsImpact {
		if len(contract.AffectedDocuments) != 0 {
			return fmt.Errorf("decision %q must not include document proposals", contract.Decision)
		}
		return nil
	}
	if len(contract.AffectedDocuments) == 0 {
		return fmt.Errorf("propose_docs decision requires at least one document")
	}

	seen := make(map[string]struct{}, len(contract.AffectedDocuments))
	for _, proposal := range contract.AffectedDocuments {
		if !mapping.IsAllowedTarget(proposal.DocumentTarget) {
			return fmt.Errorf("document %q is not an allowed mapped target", proposal.Path)
		}
		if _, exists := seen[proposal.Path]; exists {
			return fmt.Errorf("duplicate document proposal for %q", proposal.Path)
		}
		seen[proposal.Path] = struct{}{}
		if proposal.Markdown == "" {
			return fmt.Errorf("document %q has empty markdown", proposal.Path)
		}
		if err := validateTemplateHeadings(repoRoot, proposal.Template, proposal.Markdown); err != nil {
			return fmt.Errorf("document %q: %w", proposal.Path, err)
		}
	}
	return nil
}

func validateEvidenceReferences(references []string, evidence map[string]Evidence, owner string) error {
	if len(references) == 0 {
		return fmt.Errorf("%s must cite at least one evidence id", owner)
	}
	for _, reference := range references {
		if _, ok := evidence[reference]; !ok {
			return fmt.Errorf("%s references unknown evidence id %q", owner, reference)
		}
	}
	return nil
}

func validateTemplateHeadings(repoRoot, templatePath, markdown string) error {
	cleanPath := filepath.Clean(templatePath)
	if filepath.IsAbs(cleanPath) || cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return fmt.Errorf("template path must stay within repository")
	}
	template, err := os.ReadFile(filepath.Join(repoRoot, cleanPath))
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}
	required := headings(string(template))
	actual := make(map[string]struct{})
	for _, heading := range headings(markdown) {
		actual[heading] = struct{}{}
	}
	for _, heading := range required {
		if _, ok := actual[heading]; !ok {
			return fmt.Errorf("required template heading %q is missing", heading)
		}
	}
	return nil
}

func headings(markdown string) []string {
	var result []string
	for _, line := range strings.Split(markdown, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## ") {
			result = append(result, strings.TrimSpace(strings.TrimPrefix(line, "## ")))
		}
	}
	return result
}
