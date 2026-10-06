package workflow

import (
	"fmt"
	"path"
	"strings"

	"github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/documents"
	gh "github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/github"
)

// maxReportChars keeps generated Markdown below GitHub's 65,536 character
// limit for PR bodies and comments.
const maxReportChars = 60000

// featureTitle turns "log_collection" into "log collection".
func featureTitle(feature string) string {
	if feature == "" {
		return "documentation"
	}
	return strings.ReplaceAll(feature, "_", " ")
}

// evidenceLabels maps evidence IDs to short readable labels such as
// "CODE config.yaml" or "VI #2".
func evidenceLabels(evidence []documents.Evidence) map[string]string {
	labels := make(map[string]string, len(evidence))
	for _, item := range evidence {
		switch {
		case strings.HasPrefix(item.ID, "VI-"):
			labels[item.ID] = "VI #" + strings.TrimPrefix(item.ID, "VI-")
		case strings.HasPrefix(item.ID, "PR-"):
			labels[item.ID] = "PR #" + strings.TrimPrefix(item.ID, "PR-")
		default:
			source := item.Source
			if at := strings.Index(source, "@"); at >= 0 {
				source = source[:at]
			}
			name := path.Base(source)
			if strings.HasPrefix(item.ID, "DIFF-") {
				name += " (diff)"
			}
			labels[item.ID] = item.Type + " " + name
		}
	}
	return labels
}

func labelList(ids []string, labels map[string]string) string {
	var out []string
	seen := make(map[string]struct{})
	for _, id := range ids {
		label, ok := labels[id]
		if !ok {
			label = id
		}
		if _, dup := seen[label]; dup {
			continue
		}
		seen[label] = struct{}{}
		out = append(out, "`"+label+"`")
	}
	return strings.Join(out, ", ")
}

func cell(text string) string {
	text = strings.ReplaceAll(text, "|", "\\|")
	return strings.Join(strings.Fields(text), " ")
}

func clip(report string) string {
	if len(report) <= maxReportChars {
		return report
	}
	return report[:maxReportChars] + "\n\n_Report truncated. See the workflow artifact for the full Doc Contract._\n"
}

func details(summary string, body string) string {
	return fmt.Sprintf("<details>\n<summary>%s</summary>\n\n%s\n</details>\n\n", summary, strings.TrimRight(body, "\n"))
}

func provenance(contract documents.Contract, gateway string) string {
	text := fmt.Sprintf("Resolved by Claude (`%s`)", contract.Provider.Model)
	if gateway != "" {
		text += " via `" + gateway + "`"
	}
	return text
}

func contextLine(measurement ContextMeasurement) string {
	if measurement.RepositoryFiles == 0 {
		return ""
	}
	return fmt.Sprintf("Context: %d of %d repository files read (~%.0f%% less than the full candidate context, estimated)", measurement.SelectedFiles, measurement.RepositoryFiles, measurement.ReductionPct)
}

func documentTable(contract documents.Contract) string {
	var out strings.Builder
	out.WriteString("| Document | Audience | Why selected |\n|---|---|---|\n")
	for _, proposal := range contract.AffectedDocuments {
		fmt.Fprintf(&out, "| `%s` | %s | %s |\n", proposal.Path, cell(proposal.Audience), cell(proposal.Reason))
	}
	return out.String()
}

func contextChecklist(contract documents.Contract) string {
	var out strings.Builder
	item := func(ok bool, text, tag string) {
		mark := "✓"
		if !ok {
			mark = "?"
		}
		fmt.Fprintf(&out, "- %s %s `[%s]`\n", mark, text, tag)
	}
	if contract.WorkItem.UserGoal != "" {
		item(true, "Customer goal", "VI")
	}
	if contract.WorkItem.ExpectedOutcome != "" {
		item(true, "Expected outcome", "VI")
	}
	if len(contract.WorkItem.Limitations) > 0 {
		item(true, "Limitations", "VI")
	}
	item(true, "Configuration and behavior", "CODE")
	for _, missing := range contract.MissingInformation {
		item(false, cell(missing), "MISSING")
	}
	return out.String()
}

func settingsTable(contract documents.Contract) string {
	if len(contract.Change.Settings) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("| Setting | Before | After |\n|---|---|---|\n")
	for _, setting := range contract.Change.Settings {
		fmt.Fprintf(&out, "| `%s` | %s | %s |\n", setting.Name, cell(setting.Old), cell(setting.New))
	}
	return out.String()
}

func claimsSection(contract documents.Contract, labels map[string]string) string {
	var out strings.Builder
	for _, audience := range []string{"customer", "internal-developer"} {
		var lines []string
		for _, claim := range contract.Claims {
			if claim.Audience == audience {
				lines = append(lines, fmt.Sprintf("- %s — %s", strings.TrimSpace(claim.Text), labelList(claim.EvidenceIDs, labels)))
			}
		}
		if len(lines) > 0 {
			fmt.Fprintf(&out, "**%s**\n\n%s\n\n", audience, strings.Join(lines, "\n"))
		}
	}
	return out.String()
}

func conflictsSection(conflicts []documents.Conflict, labels map[string]string) string {
	var out strings.Builder
	for _, conflict := range conflicts {
		fmt.Fprintf(&out, "- %s — %s\n", strings.TrimSpace(conflict.Description), labelList(conflict.EvidenceIDs, labels))
	}
	return out.String()
}

func blockingAlert(contract documents.Contract) string {
	var out strings.Builder
	for _, conflict := range contract.Conflicts {
		if conflict.BlocksResolution {
			fmt.Fprintf(&out, "> [!WARNING]\n> **Source conflict:** %s\n\n", strings.Join(strings.Fields(conflict.Description), " "))
		}
	}
	return out.String()
}

func validationBlock(contract documents.Contract, measurement ContextMeasurement) string {
	var out strings.Builder
	out.WriteString("### Validation\n\n")
	out.WriteString("✓ Doc Contract shape · ✓ every claim cites known evidence · ✓ only mapped documents · ✓ required template headings\n\n")
	if line := contextLine(measurement); line != "" {
		out.WriteString(line + "\n\n")
	}
	return out.String()
}

// docsPRBody renders the human-review documentation PR description.
func docsPRBody(contract documents.Contract, implementation gh.PullRequest, measurement ContextMeasurement, gateway string) string {
	labels := evidenceLabels(contract.Evidence)
	var body strings.Builder
	title := strings.Join(strings.Fields(strings.TrimPrefix(contract.WorkItem.Title, "VI:")), " ")
	fmt.Fprintf(&body, "## Documentation update: VI #%d — %s\n\n", contract.WorkItem.Number, title)
	fmt.Fprintf(&body, "Triggered by merged PR #%d · %s\n\n", implementation.Number, provenance(contract, gateway))
	body.WriteString(blockingAlert(contract))
	if contract.WorkItem.ExpectedOutcome != "" {
		fmt.Fprintf(&body, "**Expected customer outcome:** %s\n\n", cell(contract.WorkItem.ExpectedOutcome))
	}
	body.WriteString(documentTable(contract) + "\n")
	if table := settingsTable(contract); table != "" {
		body.WriteString("### What changed\n\n" + table + "\n")
	}
	body.WriteString("### Context check\n\n" + contextChecklist(contract) + "\n")
	body.WriteString(validationBlock(contract, measurement))
	if claims := claimsSection(contract, labels); claims != "" {
		body.WriteString(details(fmt.Sprintf("Evidence (%d claims)", len(contract.Claims)), claims))
	}
	var notes string
	for _, conflict := range contract.Conflicts {
		if !conflict.BlocksResolution {
			notes += conflictsSection([]documents.Conflict{conflict}, labels)
		}
	}
	if notes != "" {
		body.WriteString(details("Conflicts and notes", notes))
	}
	body.WriteString("Human review is required. This PR is never merged automatically.\n")
	return clip(body.String())
}

// resolutionComment renders the comment that blocks a VI.
func resolutionComment(contract documents.Contract) string {
	labels := evidenceLabels(contract.Evidence)
	var comment strings.Builder
	comment.WriteString("## Needs clarification\n\n")
	comment.WriteString("The documentation agent stopped: the VI and the merged product behavior disagree, or required information is missing. No docs PR was created.\n\n")
	if len(contract.Conflicts) > 0 {
		comment.WriteString("| Source conflict | Evidence |\n|---|---|\n")
		for _, conflict := range contract.Conflicts {
			fmt.Fprintf(&comment, "| %s | %s |\n", cell(conflict.Description), labelList(conflict.EvidenceIDs, labels))
		}
		comment.WriteString("\n")
	}
	if len(contract.MissingInformation) > 0 {
		comment.WriteString("### Question for a human\n\n")
		for _, missing := range contract.MissingInformation {
			fmt.Fprintf(&comment, "- %s\n", strings.Join(strings.Fields(missing), " "))
		}
		comment.WriteString("\n")
	}
	comment.WriteString("Correct the VI or the product change, then re-apply `vi:ready-for-docs`.\n")
	return clip(comment.String())
}

// renderSummary renders the GitHub Actions job summary.
func renderSummary(title string, contract documents.Contract, issue gh.Issue, pull gh.PullRequest, measurement ContextMeasurement, gateway, docsPR string) string {
	labels := evidenceLabels(contract.Evidence)
	var out strings.Builder
	fmt.Fprintf(&out, "## %s\n\n", title)
	fmt.Fprintf(&out, "VI [#%d](%s) · implementation PR [#%d](%s) · %s\n\n", issue.Number, issue.HTMLURL, pull.Number, pull.HTMLURL, provenance(contract, gateway))
	fmt.Fprintf(&out, "**Decision:** `%s`", contract.Decision)
	if docsPR != "" {
		fmt.Fprintf(&out, " · **Docs PR:** %s", docsPR)
	}
	out.WriteString("\n\n")
	out.WriteString(blockingAlert(contract))
	if len(contract.AffectedDocuments) > 0 {
		out.WriteString(documentTable(contract) + "\n")
	}
	if line := contextLine(measurement); line != "" {
		out.WriteString(line + "\n\n")
	}
	if len(contract.Conflicts) > 0 {
		out.WriteString(details("Conflicts", conflictsSection(contract.Conflicts, labels)))
	}
	if len(contract.MissingInformation) > 0 {
		var list strings.Builder
		for _, item := range contract.MissingInformation {
			fmt.Fprintf(&list, "- %s\n", strings.Join(strings.Fields(item), " "))
		}
		out.WriteString(details("Missing information", list.String()))
	}
	out.WriteString("Human review is required; documentation is never auto-merged.\n")
	return clip(out.String())
}
