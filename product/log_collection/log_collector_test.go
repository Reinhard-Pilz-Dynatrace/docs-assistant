package logcollection

import (
	"os"
	"testing"
)

func readCheckedInConfig(t *testing.T) ProductConfig {
	t.Helper()
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		t.Fatalf("read checked-in config: %v", err)
	}
	config, err := ParseConfig(data)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	return config
}

func TestCheckedInConfigDefaults(t *testing.T) {
	config := readCheckedInConfig(t).LogCollection
	if !config.Enabled || config.MaxFileSizeMB != 100 || config.IncludeRotatedFiles {
		t.Fatalf("checked-in defaults = %#v", config)
	}
}

func TestParseConfigRejectsInvalidMaxFileSize(t *testing.T) {
	if _, err := ParseConfig([]byte("log_collection:\n  enabled: true\n  max_file_size_mb: 0\n")); err == nil {
		t.Fatal("ParseConfig() accepted a zero max_file_size_mb")
	}
}

func TestShouldCollect(t *testing.T) {
	config := Config{Enabled: true, MaxFileSizeMB: 100}
	tests := []struct {
		name    string
		config  Config
		sizeMB  int
		rotated bool
		want    bool
	}{
		{"small active file", config, 10, false, true},
		{"file at the limit", config, 100, false, true},
		{"file over the limit", config, 101, false, false},
		{"rotated file excluded by default", config, 10, true, false},
		{"rotated file included when configured", Config{Enabled: true, MaxFileSizeMB: 100, IncludeRotatedFiles: true}, 10, true, true},
		{"disabled collection", Config{MaxFileSizeMB: 100}, 10, false, false},
	}
	for _, test := range tests {
		if got := ShouldCollect(test.config, test.sizeMB, test.rotated); got != test.want {
			t.Errorf("%s: ShouldCollect() = %t, want %t", test.name, got, test.want)
		}
	}
}
