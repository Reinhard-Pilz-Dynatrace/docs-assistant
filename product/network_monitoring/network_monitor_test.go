package networkmonitoring

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
	got := config.NetworkMonitoring
	if got.Enabled || got.SampleRatePercent != 10 || got.CaptureDNSQueries {
		t.Fatalf("checked-in defaults = %#v", got)
	}
}

func TestParseConfigRejectsOutOfRangeSampleRate(t *testing.T) {
	for _, rate := range []string{"0", "101"} {
		if _, err := ParseConfig([]byte("network_monitoring:\n  sample_rate_percent: " + rate + "\n")); err == nil {
			t.Errorf("ParseConfig() accepted sample_rate_percent %s", rate)
		}
	}
}

func TestShouldSample(t *testing.T) {
	config := Config{Enabled: true, SampleRatePercent: 10}
	if !ShouldSample(config, 10) || ShouldSample(config, 11) {
		t.Fatal("expected every 10th connection to be sampled")
	}
	if ShouldSample(Config{SampleRatePercent: 10}, 10) {
		t.Fatal("disabled monitoring must not sample")
	}
}

func TestShouldCaptureDNS(t *testing.T) {
	if ShouldCaptureDNS(Config{Enabled: true}) || !ShouldCaptureDNS(Config{Enabled: true, CaptureDNSQueries: true}) || ShouldCaptureDNS(Config{CaptureDNSQueries: true}) {
		t.Fatal("DNS capture requires both enabled and capture_dns_queries")
	}
}
