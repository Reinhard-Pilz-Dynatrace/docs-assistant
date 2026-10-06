package errortracking

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
	got := config.ErrorTracking
	if got.Enabled != true || got.MaxStackFrames != 50 || got.GroupSimilarErrors != false {
		t.Fatalf("checked-in defaults = %#v", got)
	}
}

func TestParseConfigRejectsOutOfRangeMaxStackFrames(t *testing.T) {
	for _, value := range []string{"0", "501"} {
		if _, err := ParseConfig([]byte("error_tracking:\n  max_stack_frames: " + value + "\n")); err == nil {
			t.Errorf("ParseConfig() accepted max_stack_frames %s", value)
		}
	}
}

func TestShouldKeepFrame(t *testing.T) {
	config := Config{Enabled: true, MaxStackFrames: 50}
	if !ShouldKeepFrame(config, 49) || ShouldKeepFrame(config, 50) {
		t.Fatal("expected frames below the limit to be kept")
	}
	if ShouldKeepFrame(Config{MaxStackFrames: 50}, 0) {
		t.Fatal("disabled tracking must not keep frames")
	}
}

func TestShouldGroupSimilar(t *testing.T) {
	if !ShouldGroupSimilar(Config{Enabled: true, GroupSimilarErrors: true}) || ShouldGroupSimilar(Config{Enabled: true}) || ShouldGroupSimilar(Config{GroupSimilarErrors: true}) {
		t.Fatal("grouping requires both enabled and group_similar_errors")
	}
}
