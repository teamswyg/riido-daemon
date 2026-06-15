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

// An override set to a non-existent path must fail closed, NOT fall
// back to PATH lookup. Otherwise a misconfigured RIIDO_*_PATH would
// silently run a different binary than the operator chose.
func TestResolveExecutableMissingOverrideFailsClosed(t *testing.T) {
	got, ok := ResolveExecutable("sh", "/definitely/not/real-xyz")
	if ok {
		t.Fatalf("override pointing nowhere must NOT fall back to PATH, got %q", got)
	}
}

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

func TestResolveExecutableMissing(t *testing.T) {
	_, ok := ResolveExecutable("definitely-not-a-real-binary-xyz", "")
	if ok {
		t.Fatal("expected not found")
	}
}

// A GUI/launchd-spawned daemon inherits a minimal process PATH that omits
// the directory where the provider CLI is installed. Resolution must still
// find the binary through the augmented search dirs (login-shell PATH and
// well-known install locations).
func TestResolveExecutableFindsToolOutsideProcessPATH(t *testing.T) {
	minimalDir := t.TempDir()
	toolDir := t.TempDir()
	tool := writeExecutable(t, filepath.Join(toolDir, "fake-claude"), "v1")
	t.Setenv("PATH", minimalDir)

	restore := augmentedSearchDirs
	augmentedSearchDirs = func() []string { return []string{minimalDir, toolDir} }
	t.Cleanup(func() { augmentedSearchDirs = restore })

	got, ok := ResolveExecutable("fake-claude", "")
	if !ok || got != tool {
		t.Fatalf("want %s found via augmented dirs, got %q ok=%v", tool, got, ok)
	}
}

func TestLaunchPATHUsesAugmentedSearchDirs(t *testing.T) {
	firstDir := t.TempDir()
	secondDir := t.TempDir()
	restore := augmentedSearchDirs
	augmentedSearchDirs = func() []string { return []string{firstDir, secondDir} }
	t.Cleanup(func() { augmentedSearchDirs = restore })

	got := LaunchPATH()
	want := firstDir + string(os.PathListSeparator) + secondDir
	if got != want {
		t.Fatalf("LaunchPATH = %q, want %q", got, want)
	}
}

func TestEnvMapWithLaunchPATHAddsFrozenPath(t *testing.T) {
	binDir := t.TempDir()
	restore := augmentedSearchDirs
	augmentedSearchDirs = func() []string { return []string{binDir} }
	t.Cleanup(func() { augmentedSearchDirs = restore })

	got := EnvMapWithLaunchPATH(map[string]string{"RIIDO_TEST": "1"})
	if got["RIIDO_TEST"] != "1" {
		t.Fatalf("existing env value missing: %+v", got)
	}
	if path := EnvMapPATHValue(got); path != binDir {
		t.Fatalf("PATH = %q, want %q", path, binDir)
	}
}

func TestEnvMapWithLaunchPATHPreservesExplicitPath(t *testing.T) {
	got := EnvMapWithLaunchPATH(map[string]string{pathEnvKey(): "/custom/bin"})
	if path := EnvMapPATHValue(got); path != "/custom/bin" {
		t.Fatalf("PATH = %q, want explicit value", path)
	}
}

func TestEnvListWithLaunchPATHFromMapUsesFrozenPath(t *testing.T) {
	got := EnvListWithLaunchPATHFromMap(
		[]string{"RIIDO_TEST=1"},
		map[string]string{pathEnvKey(): "/frozen/bin"},
	)
	path, ok := envListValue(got, pathEnvKey())
	if !ok || path != "/frozen/bin" {
		t.Fatalf("spawn PATH = %q ok=%v, env=%v", path, ok, got)
	}
}

func TestEnvListWithLaunchPATHFromMapPreservesSpawnPath(t *testing.T) {
	got := EnvListWithLaunchPATHFromMap(
		[]string{pathEnvKey() + "=/spawn/bin"},
		map[string]string{pathEnvKey(): "/frozen/bin"},
	)
	path, ok := envListValue(got, pathEnvKey())
	if !ok || path != "/spawn/bin" {
		t.Fatalf("spawn PATH = %q ok=%v, env=%v", path, ok, got)
	}
}

func TestEnvListWithLaunchPATHFromMapPreservesExplicitEmptyPath(t *testing.T) {
	got := EnvListWithLaunchPATHFromMap(nil, map[string]string{pathEnvKey(): ""})
	path, ok := envListValue(got, pathEnvKey())
	if !ok || path != "" {
		t.Fatalf("spawn PATH = %q ok=%v, env=%v", path, ok, got)
	}
}
