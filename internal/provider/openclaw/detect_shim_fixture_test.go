package openclaw

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func writeShim(t *testing.T, version string) string {
	t.Helper()
	dir := t.TempDir()
	return writeShimInDir(t, dir, version)
}

func writeShimInDir(t *testing.T, dir, version string) string {
	t.Helper()
	path := filepath.Join(dir, "openclaw")
	script := "#!/bin/sh\necho '" + version + "'\nexit 0\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write shim: %v", err)
	}
	return path
}

func writeShimFromFixture(t *testing.T, fixture string, exitCode int) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", fixture))
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixture, err)
	}
	dir := t.TempDir()
	contentPath := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(contentPath, body, 0o644); err != nil {
		t.Fatalf("write content: %v", err)
	}
	exePath := filepath.Join(dir, "openclaw")
	script := "#!/bin/sh\ncat " + contentPath + "\nexit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(exePath, []byte(script), 0o755); err != nil {
		t.Fatalf("write shim: %v", err)
	}
	return exePath
}
