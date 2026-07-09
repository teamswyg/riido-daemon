package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallFixtureCreatesAssetsAndCleanupRemovesRoot(t *testing.T) {
	fixture, cleanup, err := newInstallFixture()
	if err != nil {
		t.Fatalf("new fixture: %v", err)
	}
	root := filepath.Dir(fixture.assetDir)
	for _, path := range []string{
		filepath.Join(fixture.assetDir, "riido-daemon_darwin_arm64.tar.gz"),
		filepath.Join(fixture.assetDir, "SHA256SUMS"),
		filepath.Join(fixture.binDir, "curl"),
		filepath.Join(fixture.binDir, "install"),
		filepath.Join(fixture.binDir, "uname"),
	} {
		if _, err := os.Stat(path); err != nil {
			cleanup()
			t.Fatalf("fixture file missing %s: %v", path, err)
		}
	}
	cleanup()
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("cleanup should remove root, err=%v", err)
	}
}

func TestWriteFixtureFilesReportsDirectoryConflict(t *testing.T) {
	root := t.TempDir()
	assetFile := filepath.Join(root, "asset-file")
	if err := os.WriteFile(assetFile, []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	fixture := installFixture{
		assetDir:   filepath.Join(assetFile, "nested"),
		binDir:     filepath.Join(root, "bin"),
		installDir: filepath.Join(root, "install"),
		marker:     filepath.Join(root, "marker"),
	}
	if err := writeFixtureFiles(fixture); err == nil {
		t.Fatalf("expected directory conflict")
	}
}

func TestOutputPathAndWriteJSONCreateEvidenceFile(t *testing.T) {
	root := t.TempDir()
	absolute := filepath.Join(root, "absolute.json")
	if got := outputPath("ignored", absolute); got != absolute {
		t.Fatalf("absolute output path changed: %q", got)
	}
	relative := outputPath(root, "nested/evidence.json")
	if relative != filepath.Join(root, "nested", "evidence.json") {
		t.Fatalf("relative output path=%q", relative)
	}
	if err := writeJSON(relative, evidenceFile{SchemaVersion: "test.v1", Status: statusPassed}); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}
	data, err := os.ReadFile(relative)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if !strings.HasSuffix(string(data), "\n") || !strings.Contains(string(data), `"status": "passed"`) {
		t.Fatalf("unexpected evidence JSON: %s", data)
	}
}
