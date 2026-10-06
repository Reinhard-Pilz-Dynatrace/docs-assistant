package metricexport

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ProductConfig struct {
	MetricExport Config `yaml:"metric_export"`
}

type Config struct {
	Enabled               bool `yaml:"enabled"`
	ExportIntervalSeconds int  `yaml:"export_interval_seconds"`
	IncludeHistograms     bool `yaml:"include_histograms"`
}

func ParseConfig(data []byte) (ProductConfig, error) {
	var config ProductConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return ProductConfig{}, fmt.Errorf("parse product config: %w", err)
	}
	value := config.MetricExport.ExportIntervalSeconds
	if value < 1 || value > 3600 {
		return ProductConfig{}, fmt.Errorf("export_interval_seconds must be between 1 and 3600")
	}
	return config, nil
}

// ShouldExport reports whether an export is due.
func ShouldExport(config Config, secondsSinceLastExport int) bool {
	return config.Enabled && secondsSinceLastExport >= config.ExportIntervalSeconds
}

// ShouldIncludeHistograms reports whether histogram metrics are part of an export.
func ShouldIncludeHistograms(config Config) bool {
	return config.Enabled && config.IncludeHistograms
}
