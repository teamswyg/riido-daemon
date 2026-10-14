package detectutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveExecutableEnvOverrideWins(t *testing.T) {
	override := filepath.Join(t.TempDir(), "fake-claude")
	if err := os.WriteFile(override, []byte("#!/bin/sh\necho hi\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, ok := ResolveExecutable("claude", override)
	if !ok || got != override {
		t.Fatalf("override should win: %q ok=%v", got, ok)
	}
}

func TestResolveExecutableMissingOverrideFailsClosed(t *testing.T) {
	got, ok := ResolveExecutable("sh", "/definitely/not/real-xyz")
	if ok {
		t.Fatalf("override pointing nowhere must NOT fall back to PATH, got %q", got)
	}
}

func TestResolveExecutableCandidatesOverrideIsOnlyCandidate(t *testing.T) {
	override := writeExecutable(t, filepath.Join(t.TempDir(), "fake-tool"), "override")
	pathDir := t.TempDir()
	_ = writeExecutable(t, filepath.Join(pathDir, "fake-tool"), "path")
	t.Setenv("PATH", pathDir)

	got := ResolveExecutableCandidates("fake-tool", override)
	if len(got) != 1 || got[0] != override {
		t.Fatalf("override must be the only candidate, got %v", got)
	}
}

func TestResolveExecutableCandidatesMissingOverrideFailsClosed(t *testing.T) {
	pathDir := t.TempDir()
	_ = writeExecutable(t, filepath.Join(pathDir, "fake-tool"), "path")
	t.Setenv("PATH", pathDir)

	got := ResolveExecutableCandidates("fake-tool", "/definitely/not/real-xyz")
	if len(got) != 0 {
		t.Fatalf("missing override must not fall back to PATH, got %v", got)
	}
}
