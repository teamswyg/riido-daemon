package main

import (
	"os"
	"path/filepath"
	"testing"
)

func localDaemonManifest() Manifest {
	return Manifest{
		SchemaVersion: "v", ID: "local", Title: "Local", GeneratedDoc: "doc.md",
		Workflow: "wf.yml", EvidenceArtifact: "evidence.json", Purpose: "purpose",
		Facts:        []Fact{{Name: "fact", Summary: "summary", SourceChecks: []string{"src"}}},
		Boundaries:   []Boundary{{Name: "daemon", Owner: "local", Summary: "summary"}},
		SourceChecks: []SourceCheck{{Name: "src", File: "src.go", Contains: "needle"}},
		Assertions:   []string{"assertion"},
	}
}

func mustLocalDaemonWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
