package metricexport

import (
	"os"
	"testing"
)

func TestCheckedInConfigDefaults(t *testing.T) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		t.Fatalf("read checked-in config: %v", err)
	}
	config, err := ParseConfig(data)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	got := config.MetricExport
	if got.Enabled != true || got.ExportIntervalSeconds != 15 || got.IncludeHistograms != true {
		t.Fatalf("checked-in defaults = %#v", got)
	}
}

func TestParseConfigRejectsOutOfRangeExportIntervalSeconds(t *testing.T) {
	for _, value := range []string{"0", "3601"} {
		if _, err := ParseConfig([]byte("metric_export:\n  export_interval_seconds: " + value + "\n")); err == nil {
			t.Errorf("ParseConfig() accepted export_interval_seconds %s", value)
		}
	}
}

func TestShouldExport(t *testing.T) {
	config := Config{Enabled: true, ExportIntervalSeconds: 15}
	if !ShouldExport(config, 15) || ShouldExport(config, 14) {
		t.Fatal("expected an export once the interval has passed")
	}
	if ShouldExport(Config{ExportIntervalSeconds: 15}, 60) {
		t.Fatal("disabled export must not export")
	}
}

func TestShouldIncludeHistograms(t *testing.T) {
	if !ShouldIncludeHistograms(Config{Enabled: true, IncludeHistograms: true}) || ShouldIncludeHistograms(Config{Enabled: true}) || ShouldIncludeHistograms(Config{IncludeHistograms: true}) {
		t.Fatal("histograms require both enabled and include_histograms")
	}
}
