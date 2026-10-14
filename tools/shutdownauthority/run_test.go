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

func TestRunDetectsTimeoutDrift(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	writeFile(t, repo, "pkg/lifecycle/shutdown.go", "package lifecycle\nconst DefaultForcedShutdownTimeout = 2 * time.Second\n")
	err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest})
	if err == nil {
		t.Fatal("expected timeout drift")
	}
}
