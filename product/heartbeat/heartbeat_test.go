package heartbeat

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
	got := config.Heartbeat
	if got.Enabled != true || got.HeartbeatIntervalSeconds != 20 || got.IncludeHostMetadata != true {
		t.Fatalf("checked-in defaults = %#v", got)
	}
}

func TestParseConfigRejectsOutOfRangeHeartbeatIntervalSeconds(t *testing.T) {
	for _, value := range []string{"4", "601"} {
		if _, err := ParseConfig([]byte("heartbeat:\n  heartbeat_interval_seconds: " + value + "\n")); err == nil {
			t.Errorf("ParseConfig() accepted heartbeat_interval_seconds %s", value)
		}
	}
}

func TestIsHeartbeatDue(t *testing.T) {
	config := Config{Enabled: true, HeartbeatIntervalSeconds: 20}
	if !IsHeartbeatDue(config, 20) || IsHeartbeatDue(config, 19) {
		t.Fatal("expected a heartbeat once the interval has passed")
	}
	if IsHeartbeatDue(Config{HeartbeatIntervalSeconds: 20}, 60) {
		t.Fatal("disabled heartbeat must not be due")
	}
}

func TestShouldIncludeHostMetadata(t *testing.T) {
	if !ShouldIncludeHostMetadata(Config{Enabled: true, IncludeHostMetadata: true}) || ShouldIncludeHostMetadata(Config{Enabled: true}) || ShouldIncludeHostMetadata(Config{IncludeHostMetadata: true}) {
		t.Fatal("host metadata requires both enabled and include_host_metadata")
	}
}
