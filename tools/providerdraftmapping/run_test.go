package main

import (
	"context"
	"path/filepath"
	"testing"
)

func TestRunWritesEvidence(t *testing.T) {
	repo := t.TempDir()
	manifestPath := "provider-event-draft.riido.json"
	docPath := "provider-event-draft.md"
	mustWrite(t, filepath.Join(repo, "provider_event_draft.go"), sourceFixture())
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
	repo := t.TempDir()
	manifestPath := "provider-event-draft.riido.json"
	mustWrite(t, filepath.Join(repo, "provider_event_draft.go"), sourceFixture())
	mustWriteManifest(t, repo, manifestPath, validManifest("provider-event-draft.md"))
	mustWrite(t, filepath.Join(repo, "provider-event-draft.md"), "stale")
	if err := run(context.Background(), options{Repo: repo, Manifest: manifestPath, CheckDoc: true}); err == nil {
		t.Fatalf("expected doc drift failure")
	}
}
