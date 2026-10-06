package tracesampling

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
	got := config.TraceSampling
	if got.Enabled != true || got.SampleRatioPercent != 25 || got.PropagateBaggage != false {
		t.Fatalf("checked-in defaults = %#v", got)
	}
}

func TestParseConfigRejectsOutOfRangeSampleRatioPercent(t *testing.T) {
	for _, value := range []string{"0", "101"} {
		if _, err := ParseConfig([]byte("trace_sampling:\n  sample_ratio_percent: " + value + "\n")); err == nil {
			t.Errorf("ParseConfig() accepted sample_ratio_percent %s", value)
		}
	}
}

func TestShouldSample(t *testing.T) {
	config := Config{Enabled: true, SampleRatioPercent: 5}
	if !ShouldSample(config, 20) || ShouldSample(config, 21) {
		t.Fatal("expected every 20th trace to be sampled")
	}
	if ShouldSample(Config{SampleRatioPercent: 5}, 20) {
		t.Fatal("disabled sampling must not sample")
	}
}

func TestShouldPropagateBaggage(t *testing.T) {
	if ShouldPropagateBaggage(Config{Enabled: true}) || !ShouldPropagateBaggage(Config{Enabled: true, PropagateBaggage: true}) || ShouldPropagateBaggage(Config{PropagateBaggage: true}) {
		t.Fatal("baggage propagation requires both enabled and propagate_baggage")
	}
}
