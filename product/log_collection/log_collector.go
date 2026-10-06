package logcollection

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ProductConfig struct {
	LogCollection Config `yaml:"log_collection"`
}

type Config struct {
	Enabled             bool `yaml:"enabled"`
	MaxFileSizeMB       int  `yaml:"max_file_size_mb"`
	IncludeRotatedFiles bool `yaml:"include_rotated_files"`
}

func ParseConfig(data []byte) (ProductConfig, error) {
	var config ProductConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return ProductConfig{}, fmt.Errorf("parse product config: %w", err)
	}
	if config.LogCollection.MaxFileSizeMB <= 0 {
		return ProductConfig{}, fmt.Errorf("max_file_size_mb must be greater than zero")
	}
	return config, nil
}

// ShouldCollect reports whether a log file is collected. Files larger than
// MaxFileSizeMB are skipped; rotated files are collected only when configured.
func ShouldCollect(config Config, sizeMB int, rotated bool) bool {
	if !config.Enabled || sizeMB > config.MaxFileSizeMB {
		return false
	}
	return !rotated || config.IncludeRotatedFiles
}
