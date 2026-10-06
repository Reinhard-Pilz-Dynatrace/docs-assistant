package heartbeat

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ProductConfig struct {
	Heartbeat Config `yaml:"heartbeat"`
}

type Config struct {
	Enabled                  bool `yaml:"enabled"`
	HeartbeatIntervalSeconds int  `yaml:"heartbeat_interval_seconds"`
	IncludeHostMetadata      bool `yaml:"include_host_metadata"`
}

func ParseConfig(data []byte) (ProductConfig, error) {
	var config ProductConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return ProductConfig{}, fmt.Errorf("parse product config: %w", err)
	}
	value := config.Heartbeat.HeartbeatIntervalSeconds
	if value < 5 || value > 600 {
		return ProductConfig{}, fmt.Errorf("heartbeat_interval_seconds must be between 5 and 600")
	}
	return config, nil
}

// IsHeartbeatDue reports whether the next heartbeat should be sent.
func IsHeartbeatDue(config Config, secondsSinceLast int) bool {
	return config.Enabled && secondsSinceLast >= config.HeartbeatIntervalSeconds
}

// ShouldIncludeHostMetadata reports whether host metadata is attached to a heartbeat.
func ShouldIncludeHostMetadata(config Config) bool {
	return config.Enabled && config.IncludeHostMetadata
}
