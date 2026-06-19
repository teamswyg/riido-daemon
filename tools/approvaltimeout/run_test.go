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

func TestRunDetectsMissingIdleBound(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	writeFile(t, repo, "internal/agentbridge/session/session_tool_approval_resolver.go", "package session\n")
	err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest})
	if err == nil {
		t.Fatal("expected source drift")
	}
}
