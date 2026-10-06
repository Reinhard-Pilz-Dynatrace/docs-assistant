package pathutil

import "testing"

func TestNormalizeSlashes(t *testing.T) {
	if got := NormalizeSlashes(`a\b\c`); got != "a/b/c" {
		t.Fatalf("NormalizeSlashes() = %q", got)
	}
}
