package detectutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveExecutablePathFallback(t *testing.T) {
	got, ok := ResolveExecutable("sh", "")
	if !ok || got == "" {
		t.Fatalf("sh should be on PATH: %q ok=%v", got, ok)
	}
}

func TestResolveExecutableCandidatesPreservePathOrder(t *testing.T) {
	firstDir := t.TempDir()
	secondDir := t.TempDir()
	first := writeExecutable(t, filepath.Join(firstDir, "fake-tool"), "first")
	second := writeExecutable(t, filepath.Join(secondDir, "fake-tool"), "second")
	t.Setenv("PATH", firstDir+string(os.PathListSeparator)+secondDir)

	got := ResolveExecutableCandidates("fake-tool", "")
	if len(got) != 2 {
		t.Fatalf("candidate count: got %d (%v), want 2", len(got), got)
	}
	if got[0] != first || got[1] != second {
		t.Fatalf("candidate order: got %v, want [%s %s]", got, first, second)
	}
}

func TestResolveExecutableMissing(t *testing.T) {
	_, ok := ResolveExecutable("definitely-not-a-real-binary-xyz", "")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestResolveExecutableFindsToolOutsideProcessPATH(t *testing.T) {
	minimalDir := t.TempDir()
	toolDir := t.TempDir()
	tool := writeExecutable(t, filepath.Join(toolDir, "fake-claude"), "v1")
	t.Setenv("PATH", minimalDir)
	overrideAugmentedSearchDirs(t, minimalDir, toolDir)

	got, ok := ResolveExecutable("fake-claude", "")
	if !ok || got != tool {
		t.Fatalf("want %s found via augmented dirs, got %q ok=%v", tool, got, ok)
	}
}
