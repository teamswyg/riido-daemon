package main

import (
	"context"
	"path/filepath"
	"testing"
)

func TestRunWritesEvidence(t *testing.T) {
	repo := testRepo(t)
	manifestPath := "idle-watchdog.riido.json"
	docPath := "idle-watchdog.md"
	mustWriteManifest(t, repo, manifestPath, validManifest(docPath))
	evidence := filepath.Join(repo, "evidence.json")

	err := run(context.Background(), options{Repo: repo, Manifest: manifestPath, WriteDoc: true, CheckDoc: true, EvidenceOut: evidence})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !fileExists(filepath.Join(repo, docPath)) {
		t.Fatalf("generated doc missing")
	}
	if !fileExists(evidence) {
		t.Fatalf("evidence missing")
	}
}

func TestRunRejectsDocDrift(t *testing.T) {
	repo := testRepo(t)
	manifestPath := "idle-watchdog.riido.json"
	mustWriteManifest(t, repo, manifestPath, validManifest("idle-watchdog.md"))
	mustWrite(t, filepath.Join(repo, "idle-watchdog.md"), "stale")

	err := run(context.Background(), options{Repo: repo, Manifest: manifestPath, CheckDoc: true})
	if err == nil {
		t.Fatalf("expected doc drift failure")
	}
}
