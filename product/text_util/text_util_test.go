package textutil

import "testing"

func TestTruncate(t *testing.T) {
	if got := Truncate("short", 10); got != "short" {
		t.Fatalf("Truncate() = %q", got)
	}
	if got := Truncate("hello world", 5); got != "hello..." {
		t.Fatalf("Truncate() = %q", got)
	}
}
