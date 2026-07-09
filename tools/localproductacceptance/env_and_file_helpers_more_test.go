package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvHelpersPreferPresentValues(t *testing.T) {
	t.Setenv("RIIDO_TEST_PRIMARY", "primary")
	t.Setenv("RIIDO_TEST_SECONDARY", "secondary")

	if got := getenvDefault("RIIDO_TEST_PRIMARY", "fallback"); got != "primary" {
		t.Fatalf("default helper=%q", got)
	}
	if got := getenvDefault("RIIDO_TEST_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("fallback=%q", got)
	}
	if got := firstEnv("RIIDO_TEST_MISSING", "RIIDO_TEST_SECONDARY"); got != "secondary" {
		t.Fatalf("first env=%q", got)
	}
	if got := firstEnv("RIIDO_TEST_MISSING_A", "RIIDO_TEST_MISSING_B"); got != "" {
		t.Fatalf("missing first env=%q", got)
	}
}

func TestLocalFileAndCaptureDirectoryHelpers(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "capture.png")
	if err := os.WriteFile(file, []byte("png"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !localFileExists(file) {
		t.Fatalf("expected file to exist: %s", file)
	}
	if localFileExists(dir) || localFileExists("") || localFileExists(filepath.Join(dir, "missing")) {
		t.Fatal("directories, blank paths, and missing files should not count")
	}

	uploadDir := ".riido-local/screenshots"
	if !captureCoveredByUploadDir(uploadDir) || !captureCoveredByUploadDir(uploadDir+"/") {
		t.Fatalf("expected %s to cover feature capture", uploadDir)
	}
	if captureCoveredByUploadDir(".riido-local/other") {
		t.Fatal("unrelated upload dir should not cover feature capture")
	}
}
