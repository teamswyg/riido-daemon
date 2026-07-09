package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteJSONReportsEvidenceDirAndMarshalErrors(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "not-dir"), []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := writeJSON(filepath.Join(root, "not-dir", "evidence.json"), evidenceFile{})
	if err == nil || !strings.Contains(err.Error(), "create evidence dir:") {
		t.Fatalf("expected evidence dir problem, got %v", err)
	}

	err = writeJSON(filepath.Join(root, "evidence.json"), map[string]any{"bad": func() {}})
	if err == nil || !strings.Contains(err.Error(), "marshal evidence:") {
		t.Fatalf("expected marshal problem, got %v", err)
	}
}

func TestNewInstallFixtureCreatesExpectedFilesAndCleansUp(t *testing.T) {
	fixture, cleanup, err := newInstallFixture()
	if err != nil {
		t.Fatalf("new fixture: %v", err)
	}
	root := filepath.Dir(fixture.assetDir)
	defer cleanup()

	for _, path := range []string{
		filepath.Join(fixture.assetDir, "riido-daemon_darwin_arm64.tar.gz"),
		filepath.Join(fixture.assetDir, "SHA256SUMS"),
		filepath.Join(fixture.binDir, "curl"),
		filepath.Join(fixture.binDir, "install"),
		filepath.Join(fixture.binDir, "uname"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("fixture path missing %s: %v", path, err)
		}
	}

	sums, err := os.ReadFile(filepath.Join(fixture.assetDir, "SHA256SUMS"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(sums), "riido-daemon_darwin_arm64.tar.gz") {
		t.Fatalf("archive checksum missing: %s", sums)
	}
	cleanup()
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("cleanup should remove fixture root, stat err=%v", err)
	}
}
