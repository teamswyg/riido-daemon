package main

import (
	"strings"
	"testing"
)

func TestLaunchdPathIncludesUserLocalBin(t *testing.T) {
	got := launchdPath()

	if !strings.Contains(got, ".local/bin") {
		t.Fatalf("launchdPath() = %q, want user local bin", got)
	}
}
