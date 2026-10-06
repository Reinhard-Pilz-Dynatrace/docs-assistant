package main

import "testing"

func TestIssueFromDocsBranch(t *testing.T) {
	tests := []struct {
		branch string
		want   int
		ok     bool
	}{
		{branch: "docs/vi-42", want: 42, ok: true},
		{branch: "feature/vi-42", ok: false},
		{branch: "docs/vi-not-a-number", ok: false},
	}
	for _, test := range tests {
		t.Run(test.branch, func(t *testing.T) {
			got, ok := issueFromDocsBranch(test.branch)
			if got != test.want || ok != test.ok {
				t.Fatalf("issueFromDocsBranch() = (%d, %t), want (%d, %t)", got, ok, test.want, test.ok)
			}
		})
	}
}
