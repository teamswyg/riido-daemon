package detectutil

import (
	"context"
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

func TestResolveExecutableMissing(t *testing.T) {
	_, ok := ResolveExecutable("definitely-not-a-real-binary-xyz", "")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestVersionProbeEchoesOutput(t *testing.T) {
	out, ok := VersionProbe(context.Background(), "/bin/echo", "1.2.3")
	if !ok {
		t.Fatal("probe failed")
	}
	if out != "1.2.3" {
		t.Fatalf("output: %q", out)
	}
}

func TestVersionProbeMissingBinary(t *testing.T) {
	_, ok := VersionProbe(context.Background(), "", "--version")
	if ok {
		t.Fatal("empty exe should fail")
	}
	_, ok = VersionProbe(context.Background(), "/no/such/path", "--version")
	if ok {
		t.Fatal("missing binary should fail")
	}
}

func TestVersionProbeStrictReportsExitCodeAndOutput(t *testing.T) {
	res := VersionProbeStrict(context.Background(), "/bin/sh", "-c", "printf 'tool failed'; exit 7")
	if !res.OK {
		t.Fatal("strict probe should report command completion")
	}
	if res.ExitCode != 7 {
		t.Fatalf("exit code: %d", res.ExitCode)
	}
	if res.Output != "tool failed" {
		t.Fatalf("output: %q", res.Output)
	}
}

func TestVersionProbeStrictMissingBinary(t *testing.T) {
	res := VersionProbeStrict(context.Background(), "/no/such/path", "--version")
	if res.OK {
		t.Fatalf("missing binary should fail closed: %+v", res)
	}
}
