package openclaw

import "testing"

func TestParseOpenClawVersionAcceptsDateStyle(t *testing.T) {
	cases := []struct {
		in   string
		want [3]int
	}{
		{"2026.5.5", [3]int{2026, 5, 5}},
		{"v2026.5.5", [3]int{2026, 5, 5}},
		{"openclaw 2026.5.5", [3]int{2026, 5, 5}},
		{"OpenClaw version 2026.05.05", [3]int{2026, 5, 5}},
		{"openclaw version 2026.12.31", [3]int{2026, 12, 31}},
	}
	for _, tc := range cases {
		got, ok := parseVersion(tc.in)
		if !ok || got != tc.want {
			t.Fatalf("parseVersion(%q): got %v ok=%v want %v", tc.in, got, ok, tc.want)
		}
	}
}
