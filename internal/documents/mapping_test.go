package documents

import (
	"strings"
	"testing"
)

func TestLoadMappingAndSelectAudienceDocs(t *testing.T) {
	mapping, err := LoadMapping(strings.NewReader(`process_monitoring:
  schema:
    - scan_interval_seconds
  docs:
    - path: docs/customer/process-monitoring.md
      audience: customer
      type: product-guide
      template: docs/templates/customer.md
    - path: docs/internal/process-monitoring.md
      audience: internal-developer
      type: developer-guide
      template: docs/templates/internal.md
`))
	if err != nil {
		t.Fatalf("LoadMapping() error = %v", err)
	}
	targets := mapping.TargetsForConcepts([]string{"scan_interval_seconds"})
	if len(targets) != 2 {
		t.Fatalf("selected %d docs, want 2", len(targets))
	}
	if targets[0].Audience != "customer" || targets[1].Audience != "internal-developer" {
		t.Fatalf("selected audiences = %q, %q", targets[0].Audience, targets[1].Audience)
	}
	if got := mapping.TargetsForConcepts([]string{"unmapped_internal_refactor"}); len(got) != 0 {
		t.Fatalf("unmapped change selected %d docs, want none", len(got))
	}
}
