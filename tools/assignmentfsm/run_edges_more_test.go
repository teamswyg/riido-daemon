package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesEvidenceForSourceDrift(t *testing.T) {
	repo := t.TempDir()
	writeFSMFixture(t, repo, "source without expected text")
	out := filepath.Join(repo, "evidence.json")
	err := run(options{Repo: repo, Manifest: "manifest.json", EvidenceOut: out})
	if err == nil || !strings.Contains(err.Error(), "missing expected source text") {
		t.Fatalf("expected source drift, got %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "missing expected source text") {
		t.Fatalf("evidence missing source drift: %s", data)
	}
}

func TestRunWritesEvidenceForDocDrift(t *testing.T) {
	repo := t.TempDir()
	writeFSMFixture(t, repo, "needle")
	if err := run(options{Repo: repo, Manifest: "manifest.json", WriteDoc: true}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(repo, "docs", "fsm.md"), "stale")
	out := filepath.Join(repo, "evidence.json")
	err := run(options{Repo: repo, Manifest: "manifest.json", CheckDoc: true, EvidenceOut: out})
	if err == nil || !strings.Contains(err.Error(), "generated doc drift") {
		t.Fatalf("expected doc drift, got %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "generated doc drift") {
		t.Fatalf("evidence missing doc drift: %s", data)
	}
}

func TestCheckDocReportsMissingDocument(t *testing.T) {
	problems := checkDoc(options{Repo: t.TempDir(), CheckDoc: true}, "missing.md", "body")
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "missing.md") {
		t.Fatalf("missing doc problem=%#v", problems)
	}
}

func TestAssignmentFSMWriteJSONReportsEncodeError(t *testing.T) {
	err := writeJSON(filepath.Join(t.TempDir(), "out.json"), make(chan int))
	if err == nil {
		t.Fatal("encode error expected")
	}
}
