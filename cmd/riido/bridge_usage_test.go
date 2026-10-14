package main

import "testing"

func TestBridgeUsageOnUnknownSubcommand(t *testing.T) {
	err := run([]string{"bridge", "nonsense"})
	if err == nil {
		t.Fatal("expected error for unknown bridge subcommand")
	}
}
