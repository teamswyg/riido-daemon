package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanRepoPathCoversSafeAndUnsafePaths(t *testing.T) {
	got, err := cleanRepoPath("/repo", "docs/shutdown.json")
	if err != nil {
		t.Fatalf("safe path rejected: %v", err)
	}
	if got != filepath.Join("/repo", "docs", "shutdown.json") {
		t.Fatalf("path=%q", got)
	}
	for _, rel := range []string{"/tmp/shutdown.json", "../shutdown.json", "docs/../secret.json"} {
		if _, err := cleanRepoPath("/repo", rel); err == nil {
			t.Fatalf("unsafe path %q accepted", rel)
		}
	}
}

func TestWriteJSONReportsMarshalAndWriteErrors(t *testing.T) {
	root := t.TempDir()
	err := writeJSON(filepath.Join(root, "bad.json"), map[string]any{"bad": func() {}})
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("expected marshal error, got %v", err)
	}
	err = writeJSON(root, Manifest{ID: "x"})
	if err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("expected write error, got %v", err)
	}
}

func TestLoadManifestRejectsUnsafePathAndParsesValidManifest(t *testing.T) {
	root := t.TempDir()
	if _, err := loadManifest(root, "../manifest.json"); err == nil {
		t.Fatal("expected unsafe manifest path to fail")
	}
	body := `{"schema_version":"v","id":"shutdown","title":"Shutdown","generated_doc":"doc.md",` +
		`"workflow":"wf","evidence_artifact":"artifact","sources":{},` +
		`"external_signals":[],"levels":[],"timeouts":[],"consumer_requirements":[],"assertions":[]}`
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest, err := loadManifest(root, "manifest.json")
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if manifest.ID != "shutdown" {
		t.Fatalf("manifest id=%q", manifest.ID)
	}
}

func TestProblemMessagesPreservesOrder(t *testing.T) {
	got := problemMessages([]problem{{"first"}, {"second"}})
	if strings.Join(got, ",") != "first,second" {
		t.Fatalf("messages=%v", got)
	}
}
