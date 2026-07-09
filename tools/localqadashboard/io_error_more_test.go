package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDashboardInputLoadersReportReadAndParseErrors(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.json")
	if _, err := loadProviderEvidence(missing); err == nil || !strings.Contains(err.Error(), "read provider evidence") {
		t.Fatalf("expected provider read error, got %v", err)
	}
	badProvider := filepath.Join(dir, "provider.json")
	mustDashboardWrite(t, badProvider, "{")
	if _, err := loadProviderEvidence(badProvider); err == nil || !strings.Contains(err.Error(), "parse provider evidence") {
		t.Fatalf("expected provider parse error, got %v", err)
	}
	if _, err := loadCoverageManifest(missing); err == nil || !strings.Contains(err.Error(), "read coverage manifest") {
		t.Fatalf("expected manifest read error, got %v", err)
	}
	badManifest := filepath.Join(dir, "manifest.json")
	mustDashboardWrite(t, badManifest, "{")
	if _, err := loadCoverageManifest(badManifest); err == nil || !strings.Contains(err.Error(), "parse coverage manifest") {
		t.Fatalf("expected manifest parse error, got %v", err)
	}
	if _, err := loadExternalEvidence(badProvider); err == nil || !strings.Contains(err.Error(), "parse external evidence") {
		t.Fatalf("expected external parse error, got %v", err)
	}
}

func TestDashboardWritersReportDirectoryAndMarshalErrors(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "blocked")
	mustDashboardWrite(t, blocked, "file")
	if err := writeText(filepath.Join(blocked, "out.html"), "x"); err == nil ||
		!strings.Contains(err.Error(), "create dashboard dir") {
		t.Fatalf("expected text directory error, got %v", err)
	}
	if err := writeJSON(filepath.Join(blocked, "out.json"), map[string]string{"ok": "yes"}); err == nil ||
		!strings.Contains(err.Error(), "create json dir") {
		t.Fatalf("expected json directory error, got %v", err)
	}
	if err := writeJSON(filepath.Join(dir, "bad.json"), map[string]any{"bad": make(chan int)}); err == nil ||
		!strings.Contains(err.Error(), "marshal json") {
		t.Fatalf("expected json marshal error, got %v", err)
	}
}

func mustDashboardWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := writeText(path, body); err != nil {
		t.Fatal(err)
	}
}
