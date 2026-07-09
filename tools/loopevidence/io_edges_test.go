package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandLoopFilesReportsMissingAndMalformedLoopFiles(t *testing.T) {
	dir := t.TempDir()
	m := manifest{LoopFiles: []string{"loops/missing.riido.json"}}
	if _, err := expandLoopFiles(dir, m); err == nil ||
		!strings.Contains(err.Error(), "load loop file") {
		t.Fatalf("expected missing loop file error, got %v", err)
	}
	writeText(filepath.Join(dir, "loops", "bad.riido.json"), "{")
	m.LoopFiles = []string{"loops/bad.riido.json"}
	if _, err := expandLoopFiles(dir, m); err == nil {
		t.Fatal("expected malformed loop file error")
	}
}

func TestLoopEvidenceWritersReportDirectoryAndMarshalErrors(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "blocked")
	if err := writeText(blocked, "file"); err != nil {
		t.Fatal(err)
	}
	if err := writeText(filepath.Join(blocked, "doc.md"), "x"); err == nil ||
		!strings.Contains(err.Error(), "create generated doc dir") {
		t.Fatalf("expected text directory error, got %v", err)
	}
	if err := writeJSON(filepath.Join(blocked, "evidence.json"), map[string]string{"ok": "yes"}); err == nil ||
		!strings.Contains(err.Error(), "create generated doc dir") {
		t.Fatalf("expected json directory error, got %v", err)
	}
	if err := writeJSON(filepath.Join(dir, "bad.json"), map[string]any{"bad": make(chan int)}); err == nil ||
		!strings.Contains(err.Error(), "marshal evidence") {
		t.Fatalf("expected marshal error, got %v", err)
	}
}

func TestRunReturnsEvidenceWriteFailure(t *testing.T) {
	dir, manifestPath := writeManifestWithLoopFile(t)
	blocked := filepath.Join(dir, "blocked")
	if err := writeText(blocked, "file"); err != nil {
		t.Fatal(err)
	}
	err := run(options{
		Repo: dir, Manifest: manifestPath,
		EvidenceOut: filepath.Join(blocked, "evidence.json"),
	})
	if err == nil || !strings.Contains(err.Error(), "create generated doc dir") {
		t.Fatalf("expected evidence write failure, got %v", err)
	}
}
