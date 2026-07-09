package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReportsDocEvidenceAndAbsentTokenEdges(t *testing.T) {
	dir := t.TempDir()
	manifestPath := "manifest.json"
	sourcePath := filepath.Join(dir, "src", "source.go")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("package src\nconst marker = \"forbidden\""), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := fullAccessFixtureManifest("doc.md", []AbsentSurface{{
		Name: "no-forbidden", Scope: []string{"src"}, Tokens: []string{"forbidden"},
	}})
	mustFullAccessJSON(t, filepath.Join(dir, manifestPath), manifest)
	err := run(options{Repo: dir, Manifest: manifestPath, WriteDoc: true, CheckDoc: true})
	assertFullAccessError(t, err, "forbidden token")
	evidencePath := filepath.Join(dir, "evidence.json")
	err = run(options{Repo: dir, Manifest: manifestPath, EvidenceOut: evidencePath})
	assertFullAccessError(t, err, "forbidden token")
	data, readErr := os.ReadFile(evidencePath)
	if readErr != nil || !strings.Contains(string(data), "forbidden token") {
		t.Fatalf("evidence read=%v data=%s", readErr, data)
	}
	if err := os.WriteFile(filepath.Join(dir, "doc.md"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest.AbsentSurfaces = nil
	mustFullAccessJSON(t, filepath.Join(dir, manifestPath), manifest)
	assertFullAccessError(t, run(options{Repo: dir, Manifest: manifestPath, CheckDoc: true}), "generated doc drift")
	badEvidence := filepath.Join(dir, "evidence-dir")
	if err := os.Mkdir(badEvidence, 0o755); err != nil {
		t.Fatal(err)
	}
	assertFullAccessError(t, run(options{Repo: dir, Manifest: manifestPath, EvidenceOut: badEvidence}), "is a directory")
}

func fullAccessFixtureManifest(doc string, absent []AbsentSurface) Manifest {
	return Manifest{
		SchemaVersion: "riido-full-access-runtime-harness.v1", ID: "test", Title: "Test",
		GeneratedDoc: doc, Workflow: "ci.yml", EvidenceArtifact: "artifact",
		Purpose: "verify", Assertions: []string{"assert"}, AbsentSurfaces: absent,
		SourceChecks: []SourceCheck{{Name: "marker", File: "src/source.go", Contains: "marker"}},
		Facts:        []Fact{{Name: "fact", Summary: "summary", SourceChecks: []string{"marker"}}},
	}
}
