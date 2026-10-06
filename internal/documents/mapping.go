package documents

import (
	"fmt"
	"io"
	"sort"

	"gopkg.in/yaml.v3"
)

type Mapping struct {
	Features map[string]FeatureMapping `yaml:"-"`
}

type FeatureMapping struct {
	Schema []string         `yaml:"schema"`
	Source []string         `yaml:"source"`
	Docs   []DocumentTarget `yaml:"docs"`
}

type DocumentTarget struct {
	Path     string `yaml:"path" json:"path"`
	Audience string `yaml:"audience" json:"audience"`
	Type     string `yaml:"type" json:"type"`
	Template string `yaml:"template" json:"template"`
	Reason   string `yaml:"reason,omitempty" json:"reason,omitempty"`
}

func LoadMapping(reader io.Reader) (Mapping, error) {
	var features map[string]FeatureMapping
	if err := yaml.NewDecoder(reader).Decode(&features); err != nil {
		return Mapping{}, fmt.Errorf("decode docs mapping: %w", err)
	}
	if len(features) == 0 {
		return Mapping{}, fmt.Errorf("docs mapping contains no features")
	}
	for feature, mapping := range features {
		if feature == "" || len(mapping.Schema) == 0 || len(mapping.Source) == 0 || len(mapping.Docs) == 0 {
			return Mapping{}, fmt.Errorf("feature %q must define schema concepts, source files, and docs", feature)
		}
		for _, doc := range mapping.Docs {
			if doc.Path == "" || doc.Audience == "" || doc.Type == "" || doc.Template == "" {
				return Mapping{}, fmt.Errorf("feature %q has an incomplete docs target", feature)
			}
		}
	}
	return Mapping{Features: features}, nil
}

// FeaturesForConcepts returns the sorted names of features whose schema
// contains at least one of the changed concepts.
func (mapping Mapping) FeaturesForConcepts(changedConcepts []string) []string {
	changed := make(map[string]struct{}, len(changedConcepts))
	for _, concept := range changedConcepts {
		changed[concept] = struct{}{}
	}
	var features []string
	for feature, featureMapping := range mapping.Features {
		for _, concept := range featureMapping.Schema {
			if _, ok := changed[concept]; ok {
				features = append(features, feature)
				break
			}
		}
	}
	sort.Strings(features)
	return features
}

func (mapping Mapping) TargetsForConcepts(changedConcepts []string) []DocumentTarget {
	seen := make(map[string]struct{})
	var targets []DocumentTarget
	for _, feature := range mapping.FeaturesForConcepts(changedConcepts) {
		for _, target := range mapping.Features[feature].Docs {
			if _, ok := seen[target.Path]; ok {
				continue
			}
			seen[target.Path] = struct{}{}
			target.Reason = "settings-schema mapping for " + feature
			targets = append(targets, target)
		}
	}
	return targets
}

// SourceFilesForConcepts returns the de-duplicated source files of the
// features matched by the changed concepts.
func (mapping Mapping) SourceFilesForConcepts(changedConcepts []string) []string {
	seen := make(map[string]struct{})
	var files []string
	for _, feature := range mapping.FeaturesForConcepts(changedConcepts) {
		for _, path := range mapping.Features[feature].Source {
			if _, ok := seen[path]; !ok {
				seen[path] = struct{}{}
				files = append(files, path)
			}
		}
	}
	return files
}

// IsSourceFile reports whether any feature lists path as source evidence.
func (mapping Mapping) IsSourceFile(path string) bool {
	for _, feature := range mapping.Features {
		for _, source := range feature.Source {
			if source == path {
				return true
			}
		}
	}
	return false
}

// IsFeatureName reports whether name is a feature key, which doubles as the
// root key of its config file and is therefore not a changeable setting.
func (mapping Mapping) IsFeatureName(name string) bool {
	_, ok := mapping.Features[name]
	return ok
}

func (mapping Mapping) IsAllowedTarget(target DocumentTarget) bool {
	for _, feature := range mapping.Features {
		for _, candidate := range feature.Docs {
			if candidate.Path == target.Path && candidate.Audience == target.Audience && candidate.Type == target.Type && candidate.Template == target.Template {
				return true
			}
		}
	}
	return false
}
