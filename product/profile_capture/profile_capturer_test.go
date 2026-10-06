package profilecapture

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
	got := config.ProfileCapture
	if got.Enabled != false || got.MaxProfileSeconds != 120 || got.CaptureNativeFrames != false {
		t.Fatalf("checked-in defaults = %#v", got)
	}
}

func TestParseConfigRejectsOutOfRangeMaxProfileSeconds(t *testing.T) {
	for _, value := range []string{"0", "3601"} {
		if _, err := ParseConfig([]byte("profile_capture:\n  max_profile_seconds: " + value + "\n")); err == nil {
			t.Errorf("ParseConfig() accepted max_profile_seconds %s", value)
		}
	}
}

func TestShouldStopProfile(t *testing.T) {
	config := Config{Enabled: true, MaxProfileSeconds: 60}
	if !ShouldStopProfile(config, 60) || ShouldStopProfile(config, 59) {
		t.Fatal("expected the profile to stop at the limit")
	}
	if ShouldStopProfile(Config{MaxProfileSeconds: 60}, 600) {
		t.Fatal("disabled capture must not stop a profile")
	}
}

func TestShouldCaptureNativeFrames(t *testing.T) {
	if !ShouldCaptureNativeFrames(Config{Enabled: true, CaptureNativeFrames: true}) || ShouldCaptureNativeFrames(Config{Enabled: true}) || ShouldCaptureNativeFrames(Config{CaptureNativeFrames: true}) {
		t.Fatal("native frames require both enabled and capture_native_frames")
	}
}
