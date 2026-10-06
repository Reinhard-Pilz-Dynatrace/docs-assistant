package timeutil

import "testing"

func TestFormatSeconds(t *testing.T) {
	if FormatSeconds(45) != "45s" || FormatSeconds(90) != "1m30s" {
		t.Fatal("unexpected FormatSeconds output")
	}
}
