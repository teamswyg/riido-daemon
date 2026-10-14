package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIntegrationSkippedDetectsGoTestSkip(t *testing.T) {
	out := "=== RUN   TestIntegration\n--- SKIP: TestIntegration (0.01s)\nPASS\n"
	if !integrationSkipped(out) {
		t.Fatal("expected skip output to be classified as skipped")
	}
}

func TestIntegrationSkippedIgnoresPassingRun(t *testing.T) {
	out := "=== RUN   TestIntegration\n--- PASS: TestIntegration (0.01s)\nPASS\n"
	if integrationSkipped(out) {
		t.Fatal("passing integration must not be classified as skipped")
	}
}

func TestResolveExecutableUsesDaemonSearchPath(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "fake-provider")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	got, ok := resolveExecutable("fake-provider", "")
	if !ok || got != exe {
		t.Fatalf("resolveExecutable=%q ok=%v, want %q", got, ok, exe)
	}
}
