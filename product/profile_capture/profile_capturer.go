package profilecapture

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ProductConfig struct {
	ProfileCapture Config `yaml:"profile_capture"`
}

type Config struct {
	Enabled             bool `yaml:"enabled"`
	MaxProfileSeconds   int  `yaml:"max_profile_seconds"`
	CaptureNativeFrames bool `yaml:"capture_native_frames"`
}

func ParseConfig(data []byte) (ProductConfig, error) {
	var config ProductConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return ProductConfig{}, fmt.Errorf("parse product config: %w", err)
	}
	value := config.ProfileCapture.MaxProfileSeconds
	if value < 1 || value > 3600 {
		return ProductConfig{}, fmt.Errorf("max_profile_seconds must be between 1 and 3600")
	}
	return config, nil
}

// ShouldStopProfile reports whether a running profile has reached its maximum duration.
func ShouldStopProfile(config Config, elapsedSeconds int) bool {
	return config.Enabled && elapsedSeconds >= config.MaxProfileSeconds
}

// ShouldCaptureNativeFrames reports whether native stack frames are recorded.
func ShouldCaptureNativeFrames(config Config) bool {
	return config.Enabled && config.CaptureNativeFrames
}
