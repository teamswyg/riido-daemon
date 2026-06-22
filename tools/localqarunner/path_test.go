package main

import (
	"strings"
	"testing"
)

func TestLocalQAPathIncludesUserLocalBin(t *testing.T) {
	got := localQAPath()

	if !strings.Contains(got, ".local/bin") {
		t.Fatalf("localQAPath() = %q, want user local bin", got)
	}
}
