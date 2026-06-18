package main

import (
	"encoding/json"
	"testing"
)

func TestBridgeProvidersListsAllFour(t *testing.T) {
	out := captureStdout(t, func() {
		if err := run([]string{"bridge", "providers"}); err != nil {
			t.Fatalf("run: %v", err)
		}
	})

	var listing struct {
		SchemaVersion string `json:"schema_version"`
		Providers     []struct {
			Name        string   `json:"name"`
			BlockedArgs []string `json:"blocked_args"`
		} `json:"providers"`
	}
	if err := json.Unmarshal([]byte(out), &listing); err != nil {
		t.Fatalf("parse JSON %q: %v", out, err)
	}
	if listing.SchemaVersion == "" {
		t.Fatalf("schema version missing: %s", out)
	}
	want := map[string]bool{"claude": false, "codex": false, "openclaw": false, "cursor": false}
	for _, p := range listing.Providers {
		if _, ok := want[p.Name]; ok {
			want[p.Name] = true
		}
		if len(p.BlockedArgs) == 0 {
			t.Fatalf("provider %s has no blocked args", p.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("provider %s not listed in %v", name, listing.Providers)
		}
	}
}
