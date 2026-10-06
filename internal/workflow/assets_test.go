package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/documents"
)

func TestCheckedInWorkflowAndIssueFormAreValidYAML(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, path := range []string{
		filepath.Join(root, ".github", "workflows", "documentation-agent.yml"),
		filepath.Join(root, ".github", "ISSUE_TEMPLATE", "vi.yml"),
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var value map[string]any
		if err := yaml.Unmarshal(data, &value); err != nil {
			t.Errorf("parse %s: %v", path, err)
		}
	}
}

func TestCheckedInDocsMappingLoads(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "mappings", "docs-map.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	mapping, err := documents.LoadMapping(file)
	if err != nil {
		t.Fatalf("LoadMapping() error = %v", err)
	}
	targets := mapping.TargetsForConcepts([]string{"include_container_processes"})
	if len(targets) != 2 {
		t.Fatalf("mapped target count = %d, want customer and internal docs", len(targets))
	}
}
