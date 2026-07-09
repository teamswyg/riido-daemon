package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesFailedEvidenceForDocDrift(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	if err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest, WriteDoc: true}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, repo, "docs/gate.md", "stale")
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

func TestMaybeWriteDocRejectsUnsafeDocPath(t *testing.T) {
	err := maybeWriteDoc(
		options{Repo: t.TempDir(), WriteDoc: true},
		Manifest{GeneratedDoc: "../gate.md"},
		"body",
	)
	if err == nil || !strings.Contains(err.Error(), "unsafe repo path") {
		t.Fatalf("expected unsafe path error, got %v", err)
	}
}

func TestCompatibilityGateIOErrorEdges(t *testing.T) {
	repo := t.TempDir()
	_, err := loadManifest(repo, "missing.json")
	if err == nil {
		t.Fatal("missing manifest error expected")
	}
	if _, err := readSource(repo, "../outside.go"); err == nil {
		t.Fatal("unsafe source error expected")
	}
	if err := writeJSON(filepath.Join(repo, "bad.json"), make(chan int)); err == nil {
		t.Fatal("marshal error expected")
	}
	if err := writeJSON(filepath.Join(repo, "missing", "out.json"), Evidence{}); err == nil {
		t.Fatal("write error expected")
	}
}

func TestBuildEvidenceMarksProblemsFailed(t *testing.T) {
	ev := buildEvidence(
		Manifest{ID: "compatibility", EvidenceArtifact: "artifact"},
		[]problem{{Message: "source drift"}},
		[]SourceResult{{Name: "source", Pass: false}},
	)
	if ev.Status != "failed" || ev.ProblemSummaries[0] != "source drift" {
		t.Fatalf("unexpected evidence: %+v", ev)
	}
}
