package processmonitoring

import (
	"os"
	"testing"
)

func TestParseConfigReadsCheckedInDefault(t *testing.T) {
	config, err := ParseConfig([]byte("process_monitoring:\n  enabled: true\n  scan_interval_seconds: 30\n"))
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if got := config.ProcessMonitoring.ScanIntervalSeconds; got != 30 {
		t.Fatalf("scan interval = %d, want 30", got)
	}
	if config.ProcessMonitoring.IncludeContainerProcesses {
		t.Fatal("container process monitoring should default to false when absent")
	}
}

func TestParseConfigRejectsInvalidScanInterval(t *testing.T) {
	_, err := ParseConfig([]byte("process_monitoring:\n  enabled: true\n  scan_interval_seconds: 0\n"))
	if err == nil {
		t.Fatal("ParseConfig() accepted a zero scan interval")
	}
}

func TestCheckedInConfigUsesSixtySecondInterval(t *testing.T) {
	config := readCheckedInConfig(t)
	if got := config.ProcessMonitoring.ScanIntervalSeconds; got != 60 {
		t.Fatalf("checked-in scan interval = %d, want 60", got)
	}
}

func TestCheckedInConfigIncludesContainerProcesses(t *testing.T) {
	config := readCheckedInConfig(t)
	if !config.ProcessMonitoring.IncludeContainerProcesses {
		t.Fatal("container process monitoring should be enabled in the demo configuration")
	}
}

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

func TestShouldMonitorProcess(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		containerized bool
		want          bool
	}{
		{name: "host process when enabled", config: Config{Enabled: true}, want: true},
		{name: "container process excluded by default", config: Config{Enabled: true}, containerized: true, want: false},
		{name: "container process included when configured", config: Config{Enabled: true, IncludeContainerProcesses: true}, containerized: true, want: true},
		{name: "disabled monitoring", config: Config{Enabled: false, IncludeContainerProcesses: true}, containerized: true, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ShouldMonitorProcess(test.config, test.containerized); got != test.want {
				t.Fatalf("ShouldMonitorProcess() = %t, want %t", got, test.want)
			}
		})
	}
}
