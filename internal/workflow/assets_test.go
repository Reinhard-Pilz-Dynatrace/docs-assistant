package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/documents"
	gh "github.com/Reinhard-Pilz-Dynatrace/docs-assistant/internal/github"
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

func TestCheckedInMappingReferencesExistingFiles(t *testing.T) {
	root := filepath.Join("..", "..")
	file, err := os.Open(filepath.Join(root, "mappings", "docs-map.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	mapping, err := documents.LoadMapping(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(mapping.Features) < 3 {
		t.Fatalf("mapping has %d features, want at least 3", len(mapping.Features))
	}
	for feature, featureMapping := range mapping.Features {
		paths := append([]string{}, featureMapping.Source...)
		for _, doc := range featureMapping.Docs {
			paths = append(paths, doc.Path, doc.Template)
		}
		for _, path := range paths {
			if _, err := os.Stat(filepath.Join(root, path)); err != nil {
				t.Errorf("feature %s references missing file: %v", feature, err)
			}
		}
	}
}

func TestUnmappedHelperPackageSelectsNoDocs(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "mappings", "docs-map.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	mapping, err := documents.LoadMapping(file)
	if err != nil {
		t.Fatal(err)
	}
	files := []gh.PullRequestFile{{Filename: "product/text_util/text_util.go", Patch: "@@ -1 +1 @@\n-func Truncate() {}\n+func Truncate() { /* refactor */ }\n"}}
	if concepts := changedConcepts(files, mapping); len(concepts) != 0 {
		t.Fatalf("helper refactor changed concepts %#v, want none", concepts)
	}
}
