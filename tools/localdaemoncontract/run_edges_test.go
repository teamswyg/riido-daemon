package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesFailedEvidenceBeforeReturningProblems(t *testing.T) {
	dir := t.TempDir()
	m := localDaemonManifest()
	m.SourceChecks = []SourceCheck{{Name: "src", File: "src.go", Contains: "needle"}}
	writeLocalDaemonManifest(t, dir, m)
	out := filepath.Join(dir, "evidence.json")
	err := run(options{Repo: dir, Manifest: "manifest.json", EvidenceOut: out})
	if err == nil || !strings.Contains(err.Error(), "src:") {
		t.Fatalf("expected source problem, got %v", err)
	}
	var evidence Evidence
	data, readErr := os.ReadFile(out)
	if readErr != nil || json.Unmarshal(data, &evidence) != nil {
		t.Fatalf("evidence was not written: %v %s", readErr, data)
	}
	if len(evidence.Problems) == 0 || evidence.SourceChecks[0].OK {
		t.Fatalf("expected failed evidence, got %#v", evidence)
	}
}

func TestCheckDocReportsMissingAndDrift(t *testing.T) {
	dir := t.TempDir()
	opts := options{Repo: dir, CheckDoc: true}
	problems := checkDoc(opts, "doc.md", "expected")
	assertLocalDaemonProblem(t, problems, "no such file")
	mustLocalDaemonWrite(t, filepath.Join(dir, "doc.md"), "old")
	problems = checkDoc(opts, "doc.md", "expected")
	assertLocalDaemonProblem(t, problems, "generated doc drift")
}

func TestRunPropagatesWriteFailures(t *testing.T) {
	dir := t.TempDir()
	m := localDaemonManifest()
	m.GeneratedDoc = "blocked/doc.md"
	mustLocalDaemonWrite(t, filepath.Join(dir, "src.go"), "needle")
	mustLocalDaemonWrite(t, filepath.Join(dir, "blocked"), "file")
	writeLocalDaemonManifest(t, dir, m)
	if err := run(options{Repo: dir, Manifest: "manifest.json", WriteDoc: true}); err == nil {
		t.Fatal("expected doc write failure")
	}
	m.GeneratedDoc = "doc.md"
	writeLocalDaemonManifest(t, dir, m)
	if err := run(options{
		Repo: dir, Manifest: "manifest.json",
		EvidenceOut: filepath.Join(dir, "blocked", "evidence.json"),
	}); err == nil {
		t.Fatal("expected evidence write failure")
	}
}
