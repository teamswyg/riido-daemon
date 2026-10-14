package main

import "testing"

func TestDaemonStartUnknownArg(t *testing.T) {
	err := run([]string{"daemon", "start", "--bogus"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}
