package tracesampling

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ProductConfig struct {
	TraceSampling Config `yaml:"trace_sampling"`
}

type Config struct {
	Enabled            bool `yaml:"enabled"`
	SampleRatioPercent int  `yaml:"sample_ratio_percent"`
	PropagateBaggage   bool `yaml:"propagate_baggage"`
}

func ParseConfig(data []byte) (ProductConfig, error) {
	var config ProductConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return ProductConfig{}, fmt.Errorf("parse product config: %w", err)
	}
	value := config.TraceSampling.SampleRatioPercent
	if value < 1 || value > 100 {
		return ProductConfig{}, fmt.Errorf("sample_ratio_percent must be between 1 and 100")
	}
	return config, nil
}

// ShouldSample reports whether the trace with the given sequence number is
// sampled: every (100/sample_ratio_percent)-th trace, rounded down.
func ShouldSample(config Config, sequence int) bool {
	if !config.Enabled {
		return false
	}
	return sequence%(100/config.SampleRatioPercent) == 0
}

// ShouldPropagateBaggage reports whether baggage is forwarded to downstream services.
func ShouldPropagateBaggage(config Config) bool {
	return config.Enabled && config.PropagateBaggage
}
