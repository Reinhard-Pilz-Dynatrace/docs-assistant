package crashreporting

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
	got := config.CrashReporting
	if got.Enabled != true || got.MaxDumpSizeMB != 64 || got.IncludeEnvironment != false {
		t.Fatalf("checked-in defaults = %#v", got)
	}
}

func TestParseConfigRejectsOutOfRangeMaxDumpSizeMB(t *testing.T) {
	for _, value := range []string{"0", "1025"} {
		if _, err := ParseConfig([]byte("crash_reporting:\n  max_dump_size_mb: " + value + "\n")); err == nil {
			t.Errorf("ParseConfig() accepted max_dump_size_mb %s", value)
		}
	}
}

func TestShouldUploadDump(t *testing.T) {
	config := Config{Enabled: true, MaxDumpSizeMB: 64}
	if !ShouldUploadDump(config, 64) || ShouldUploadDump(config, 65) {
		t.Fatal("expected dumps up to the limit to be uploaded")
	}
	if ShouldUploadDump(Config{MaxDumpSizeMB: 64}, 1) {
		t.Fatal("disabled reporting must not upload")
	}
}

func TestShouldIncludeEnvironment(t *testing.T) {
	if !ShouldIncludeEnvironment(Config{Enabled: true, IncludeEnvironment: true}) || ShouldIncludeEnvironment(Config{Enabled: true}) || ShouldIncludeEnvironment(Config{IncludeEnvironment: true}) {
		t.Fatal("environment capture requires both enabled and include_environment")
	}
}
