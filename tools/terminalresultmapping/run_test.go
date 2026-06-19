package main

import (
	"path/filepath"
	"testing"
)

func TestRunWritesVerifiedEvidence(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	out := filepath.Join(repo, "out.json")
	err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest, WriteDoc: true, CheckDoc: true, EvidenceOut: out})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunDetectsMissingStatusMapping(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	appendFile(t, repo, "internal/agentbridge/result.go", "\nconst ResultPaused ResultStatus = \"paused\"\n")
	err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest})
	if err == nil {
		t.Fatal("expected missing status mapping failure")
	}
}
