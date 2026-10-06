package errortracking

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ProductConfig struct {
	ErrorTracking Config `yaml:"error_tracking"`
}

type Config struct {
	Enabled            bool `yaml:"enabled"`
	MaxStackFrames     int  `yaml:"max_stack_frames"`
	GroupSimilarErrors bool `yaml:"group_similar_errors"`
}

func ParseConfig(data []byte) (ProductConfig, error) {
	var config ProductConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return ProductConfig{}, fmt.Errorf("parse product config: %w", err)
	}
	value := config.ErrorTracking.MaxStackFrames
	if value < 1 || value > 500 {
		return ProductConfig{}, fmt.Errorf("max_stack_frames must be between 1 and 500")
	}
	return config, nil
}

// ShouldKeepFrame reports whether the stack frame at the given index is recorded.
func ShouldKeepFrame(config Config, frameIndex int) bool {
	return config.Enabled && frameIndex < config.MaxStackFrames
}

// ShouldGroupSimilar reports whether similar errors are merged into one entry.
func ShouldGroupSimilar(config Config) bool {
	return config.Enabled && config.GroupSimilarErrors
}
