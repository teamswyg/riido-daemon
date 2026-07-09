package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanRepoPathRejectsUnsafePaths(t *testing.T) {
	repo := t.TempDir()
	if _, err := cleanRepoPath(repo, "/tmp/outside"); err == nil {
		t.Fatal("absolute path should be rejected")
	}
	if _, err := cleanRepoPath(repo, "../outside"); err == nil {
		t.Fatal("parent path should be rejected")
	}
	got, err := cleanRepoPath(repo, "docs/out.md")
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(repo, "docs", "out.md") {
		t.Fatalf("path=%q", got)
	}
}

func TestRunWritesFailedEvidenceForDocDrift(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	if err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest, WriteDoc: true}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, repo, "docs/allowed-fields.md", "stale")
	out := filepath.Join(repo, "evidence.json")
	err := run(t.Context(), options{
		Repo: repo, Manifest: defaultManifest, CheckDoc: true, EvidenceOut: out,
	})
	if err == nil || !strings.Contains(err.Error(), "generated doc drift") {
		t.Fatalf("expected doc drift, got %v", err)
	}
	var evidence Evidence
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.Status != "failed" || len(evidence.ProblemSummaries) == 0 {
		t.Fatalf("failed evidence missing: %+v", evidence)
	}
}

func TestMaybeWriteDocsRejectsUnsafeDocPath(t *testing.T) {
	err := maybeWriteDocs(
		options{Repo: t.TempDir(), WriteDoc: true},
		Manifest{AllowedDoc: "../allowed.md", ForbiddenDoc: "forbidden.md"},
		renderedDocs{Allowed: "a", Forbidden: "f"},
	)
	if err == nil || !strings.Contains(err.Error(), "unsafe repo path") {
		t.Fatalf("expected unsafe path error, got %v", err)
	}
}

func TestWriteJSONReportsMarshalAndWriteErrors(t *testing.T) {
	err := writeJSON(filepath.Join(t.TempDir(), "bad.json"), make(chan int))
	if err == nil {
		t.Fatal("marshal error expected")
	}
	err = writeJSON(filepath.Join(t.TempDir(), "missing", "out.json"), Evidence{})
	if err == nil {
		t.Fatal("write error expected")
	}
}
