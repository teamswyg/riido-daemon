package main

import (
	"context"
	"path/filepath"
	"testing"
)

func TestRunWritesEvidence(t *testing.T) {
	repo := testRepo(t)
	manifestPath := "event-authority.riido.json"
	docPath := "event-authority.md"
	mustWriteManifest(t, repo, manifestPath, validManifest(docPath))
	evidence := filepath.Join(repo, "evidence.json")

	err := run(context.Background(), options{Repo: repo, Manifest: manifestPath, WriteDoc: true, CheckDoc: true, EvidenceOut: evidence})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !fileExists(filepath.Join(repo, docPath)) || !fileExists(evidence) {
		t.Fatalf("expected generated doc and evidence")
	}
}

func TestRunRejectsDocDrift(t *testing.T) {
	repo := testRepo(t)
	manifestPath := "event-authority.riido.json"
	mustWriteManifest(t, repo, manifestPath, validManifest("event-authority.md"))
	mustWrite(t, filepath.Join(repo, "event-authority.md"), "stale")

	if err := run(context.Background(), options{Repo: repo, Manifest: manifestPath, CheckDoc: true}); err == nil {
		t.Fatalf("expected doc drift failure")
	}
}
