package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesFailedEvidenceBeforeReturningDrift(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	writeFile(t, repo, "pkg/lifecycle/shutdown.go", "package lifecycle\nconst DefaultForcedShutdownTimeout = 2 * time.Second\n")
	out := filepath.Join(repo, "evidence.json")
	err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest, EvidenceOut: out})
	if err == nil || !strings.Contains(err.Error(), "timeout drift") {
		t.Fatalf("expected timeout drift, got %v", err)
	}
	var evidence Evidence
	data, readErr := os.ReadFile(out)
	if readErr != nil || json.Unmarshal(data, &evidence) != nil {
		t.Fatalf("failed evidence missing: err=%v body=%s", readErr, data)
	}
	if evidence.Status != "failed" || len(evidence.ProblemSummaries) == 0 {
		t.Fatalf("expected failed evidence, got %#v", evidence)
	}
}

func TestCheckDocReportsMissingDriftAndUnsafePath(t *testing.T) {
	repo := t.TempDir()
	manifest := Manifest{GeneratedDoc: "doc.md"}
	opts := options{Repo: repo, CheckDoc: true}
	problems := checkDoc(opts, manifest, "expected")
	assertShutdownProblem(t, problems, "no such file")
	writeFile(t, repo, "doc.md", "old")
	problems = checkDoc(opts, manifest, "expected")
	assertShutdownProblem(t, problems, "generated doc drift")
	manifest.GeneratedDoc = "../doc.md"
	problems = checkDoc(opts, manifest, "expected")
	assertShutdownProblem(t, problems, "unsafe repo path")
}

func TestRunPropagatesDocAndEvidenceWriteFailures(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	blocked := filepath.Join(repo, "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(repo, defaultManifest)
	body := strings.Replace(fixtureManifestSource, `"generated_doc": "docs/20-domain/provider-runtime/adapter-draft-fields/cancel-interrupt-input.md"`, `"generated_doc": "blocked/doc.md"`, 1)
	if err := os.WriteFile(manifest, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest, WriteDoc: true}); err == nil {
		t.Fatal("expected doc write failure")
	}
	writeFixtureRepo(t, repo)
	err := run(t.Context(), options{
		Repo: repo, Manifest: defaultManifest,
		EvidenceOut: filepath.Join(blocked, "evidence.json"),
	})
	if err == nil {
		t.Fatal("expected evidence write failure")
	}
}
