package networkmonitoring

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ProductConfig struct {
	NetworkMonitoring Config `yaml:"network_monitoring"`
}

type Config struct {
	Enabled           bool `yaml:"enabled"`
	SampleRatePercent int  `yaml:"sample_rate_percent"`
	CaptureDNSQueries bool `yaml:"capture_dns_queries"`
}

func ParseConfig(data []byte) (ProductConfig, error) {
	var config ProductConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return ProductConfig{}, fmt.Errorf("parse product config: %w", err)
	}
	rate := config.NetworkMonitoring.SampleRatePercent
	if rate < 1 || rate > 100 {
		return ProductConfig{}, fmt.Errorf("sample_rate_percent must be between 1 and 100")
	}
	return config, nil
}

// ShouldSample reports whether the connection with the given sequence number
// is sampled: every (100/sample_rate_percent)-th connection, rounded down.
func ShouldSample(config Config, sequence int) bool {
	if !config.Enabled {
		return false
	}
	step := 100 / config.SampleRatePercent
	return sequence%step == 0
}

// ShouldCaptureDNS reports whether DNS queries are recorded for a connection.
func ShouldCaptureDNS(config Config) bool {
	return config.Enabled && config.CaptureDNSQueries
}
