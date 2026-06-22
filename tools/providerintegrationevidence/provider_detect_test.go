package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveOpenClawExecutableUsesSupportedCandidate(t *testing.T) {
	oldDir := t.TempDir()
	newDir := t.TempDir()
	writeVersionShim(t, filepath.Join(oldDir, "openclaw"), "OpenClaw 2026.3.24")
	newExe := writeVersionShim(t, filepath.Join(newDir, "openclaw"), "OpenClaw 2026.5.22")
	t.Setenv("PATH", oldDir+string(os.PathListSeparator)+newDir)

	got, ok := resolveProviderExecutable(provider{ID: "openclaw"}, "")
	if !ok || got != newExe {
		t.Fatalf("resolveProviderExecutable=%q ok=%v, want %q", got, ok, newExe)
	}
}

func writeVersionShim(t *testing.T, path, version string) string {
	t.Helper()
	script := "#!/bin/sh\necho '" + version + "'\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
