package main

import "testing"

func TestRootHelpVisibilityContract(t *testing.T) {
	groups := []CommandGroup{
		{Name: "work", RootHelp: "public", UsageTokens: []string{"riido work get"}},
		{Name: "daemon", RootHelp: "internal", UsageTokens: []string{"riido daemon start"}},
	}
	for _, tc := range []struct {
		name, out string
		valid     bool
	}{
		{"public only", "riido work get", true},
		{"missing public", "", false},
		{"internal full", "riido work get\nriido daemon start", false},
		{"internal abbreviated", "riido work get\nriido daemon ...", false},
		{"permission helper", "riido work get\nriido claude-permission-mcp", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if detail := checkRootHelpTokens(tc.out, groups); (detail == "") != tc.valid {
				t.Fatalf("valid=%v detail=%q", tc.valid, detail)
			}
		})
	}
	for _, visibility := range []string{"", "hidden", "Public"} {
		groups[1].RootHelp = visibility
		if checkRootHelpTokens("riido work get", groups) == "" {
			t.Fatalf("accepted invalid visibility %q", visibility)
		}
	}
}
