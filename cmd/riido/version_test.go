package main

import (
	"strings"
	"testing"
)

func TestVersionCommandPrintsBinaryVersion(t *testing.T) {
	old := binaryVersion
	binaryVersion = "v-test"
	t.Cleanup(func() { binaryVersion = old })
	out := captureStdout(t, func() {
		if err := run([]string{"version"}); err != nil {
			t.Fatalf("run version: %v", err)
		}
	})
	if strings.TrimSpace(out) != "riido-daemon v-test" {
		t.Fatalf("version output = %q", out)
	}
}

func TestVersionFlagPrintsBinaryVersion(t *testing.T) {
	old := binaryVersion
	binaryVersion = "v-flag"
	t.Cleanup(func() { binaryVersion = old })
	out := captureStdout(t, func() {
		if err := run([]string{"--version"}); err != nil {
			t.Fatalf("run --version: %v", err)
		}
	})
	if strings.TrimSpace(out) != "riido-daemon v-flag" {
		t.Fatalf("--version output = %q", out)
	}
}
