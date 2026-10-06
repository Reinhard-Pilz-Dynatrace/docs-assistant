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
		if feature == "" || len(mapping.Schema) == 0 || len(mapping.Docs) == 0 {
			return Mapping{}, fmt.Errorf("feature %q must define schema concepts and docs", feature)
		}
		for _, doc := range mapping.Docs {
			if doc.Path == "" || doc.Audience == "" || doc.Type == "" || doc.Template == "" {
				return Mapping{}, fmt.Errorf("feature %q has an incomplete docs target", feature)
			}
		}
	}
	return Mapping{Features: features}, nil
}

func (mapping Mapping) TargetsForConcepts(changedConcepts []string) []DocumentTarget {
	changed := make(map[string]struct{}, len(changedConcepts))
	for _, concept := range changedConcepts {
		changed[concept] = struct{}{}
	}

	features := make([]string, 0, len(mapping.Features))
	for feature := range mapping.Features {
		features = append(features, feature)
	}
	sort.Strings(features)

	seen := make(map[string]struct{})
	var targets []DocumentTarget
	for _, feature := range features {
		featureMapping := mapping.Features[feature]
		matched := false
		for _, concept := range featureMapping.Schema {
			if _, ok := changed[concept]; ok {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		for _, target := range featureMapping.Docs {
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
