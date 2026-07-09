package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesDocAndEvidence(t *testing.T) {
	dir := t.TempDir()
	manifest := nativeManifest()
	mustNativeWrite(t, filepath.Join(dir, "src.go"), "needle")
	writeNativeManifest(t, dir, manifest)
	out := filepath.Join(dir, "out", "evidence.json")
	if err := run(options{Repo: dir, Manifest: "manifest.json", WriteDoc: true, CheckDoc: true, EvidenceOut: out}); err != nil {
		t.Fatal(err)
	}
	doc, err := os.ReadFile(filepath.Join(dir, manifest.GeneratedDoc))
	if err != nil || !strings.Contains(string(doc), manifest.Title) {
		t.Fatalf("doc=%q err=%v", doc, err)
	}
	var evidence Evidence
	data, err := os.ReadFile(out)
	if err != nil || json.Unmarshal(data, &evidence) != nil {
		t.Fatalf("evidence read/unmarshal failed: %v %s", err, data)
	}
	if evidence.ID != manifest.ID || len(evidence.SourceChecks) != 1 {
		t.Fatalf("evidence=%#v", evidence)
	}
}

func TestRunDetectsDocDrift(t *testing.T) {
	dir := t.TempDir()
	manifest := nativeManifest()
	mustNativeWrite(t, filepath.Join(dir, "src.go"), "needle")
	mustNativeWrite(t, filepath.Join(dir, manifest.GeneratedDoc), "old doc")
	writeNativeManifest(t, dir, manifest)
	err := run(options{Repo: dir, Manifest: "manifest.json", CheckDoc: true})
	if err == nil || !strings.Contains(err.Error(), "generated doc drift") {
		t.Fatalf("expected doc drift, got %v", err)
	}
}

func writeNativeManifest(t *testing.T, dir string, m Manifest) {
	t.Helper()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	mustNativeWrite(t, filepath.Join(dir, "manifest.json"), string(data))
}

func nativeManifest() Manifest {
	return Manifest{
		SchemaVersion: "v", ID: "native", Title: "Native", GeneratedDoc: "doc.md",
		Workflow: "wf.yml", EvidenceArtifact: "evidence.json", Purpose: "purpose",
		Facts:        []Fact{{Name: "fact", Summary: "summary", SourceChecks: []string{"src"}}},
		SourceChecks: []SourceCheck{{Name: "src", File: "src.go", Contains: "needle"}},
		Assertions:   []string{"assertion"},
	}
}
