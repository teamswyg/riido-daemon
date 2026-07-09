package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReportsMissingManifest(t *testing.T) {
	err := run(options{Repo: t.TempDir(), Manifest: "missing.json"})
	if err == nil || !strings.Contains(err.Error(), "read manifest") {
		t.Fatalf("expected missing manifest error, got %v", err)
	}
}

func TestRunWritesFailedEvidenceBeforeReturningProblems(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir)
	mustWrite(t, filepath.Join(dir, "src.go"), "drift")
	evidence := filepath.Join(dir, "out", "validation.json")

	err := run(options{Repo: dir, Manifest: "manifest.json", EvidenceOut: evidence})
	if err == nil || !strings.Contains(err.Error(), "source check failed") {
		t.Fatalf("expected source check error, got %v", err)
	}
	data, err := os.ReadFile(evidence)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"status": "failed"`) {
		t.Fatalf("failed evidence was not written: %s", data)
	}
}

func TestRunReturnsEvidenceWriteError(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir)
	err := run(options{
		Repo:        dir,
		Manifest:    "manifest.json",
		EvidenceOut: dir,
	})
	if err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("expected evidence write error, got %v", err)
	}
}

func TestCheckDocReportsDrift(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "validation.md"), "old")
	problems := checkDoc(
		options{Repo: dir, CheckDoc: true},
		"validation.md",
		"new",
	)
	if !hasProblem(problems, "generated doc drift") {
		t.Fatalf("expected doc drift problem, got %#v", problems)
	}
}

func TestWriteJSONReportsEncodeError(t *testing.T) {
	err := writeJSON(filepath.Join(t.TempDir(), "out.json"), make(chan int))
	if err == nil || !strings.Contains(err.Error(), "encode evidence") {
		t.Fatalf("expected encode error, got %v", err)
	}
}
