package main

import "testing"

func TestWorkspaceLabelKeepsID(t *testing.T) {
	names := map[string]string{"ws-1": "스위그"}
	for _, c := range []struct{ id, want string }{
		{"ws-1", "스위그 (ws-1)"},
		{"ws-unknown", "ws-unknown"},
		{"", ""},
	} {
		if got := workspaceLabel(c.id, names); got != c.want {
			t.Errorf("workspaceLabel(%q) = %q, want %q", c.id, got, c.want)
		}
	}
}
