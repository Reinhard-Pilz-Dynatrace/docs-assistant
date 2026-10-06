package crashreporting

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ProductConfig struct {
	CrashReporting Config `yaml:"crash_reporting"`
}

type Config struct {
	Enabled            bool `yaml:"enabled"`
	MaxDumpSizeMB      int  `yaml:"max_dump_size_mb"`
	IncludeEnvironment bool `yaml:"include_environment"`
}

func ParseConfig(data []byte) (ProductConfig, error) {
	var config ProductConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return ProductConfig{}, fmt.Errorf("parse product config: %w", err)
	}
	value := config.CrashReporting.MaxDumpSizeMB
	if value < 1 || value > 1024 {
		return ProductConfig{}, fmt.Errorf("max_dump_size_mb must be between 1 and 1024")
	}
	return config, nil
}

// ShouldUploadDump reports whether a crash dump of the given size is uploaded.
func ShouldUploadDump(config Config, sizeMB int) bool {
	return config.Enabled && sizeMB <= config.MaxDumpSizeMB
}

// ShouldIncludeEnvironment reports whether environment variables are attached to a report.
func ShouldIncludeEnvironment(config Config) bool {
	return config.Enabled && config.IncludeEnvironment
}
