package processmonitoring

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ProductConfig struct {
	ProcessMonitoring Config `yaml:"process_monitoring"`
}

type Config struct {
	Enabled                   bool `yaml:"enabled"`
	ScanIntervalSeconds       int  `yaml:"scan_interval_seconds"`
	IncludeContainerProcesses bool `yaml:"include_container_processes"`
}

func ParseConfig(data []byte) (ProductConfig, error) {
	var config ProductConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return ProductConfig{}, fmt.Errorf("parse product config: %w", err)
	}
	if config.ProcessMonitoring.ScanIntervalSeconds <= 0 {
		return ProductConfig{}, fmt.Errorf("scan_interval_seconds must be greater than zero")
	}
	return config, nil
}

func ShouldMonitorProcess(config Config, containerized bool) bool {
	if !config.Enabled {
		return false
	}
	return !containerized || config.IncludeContainerProcesses
}
