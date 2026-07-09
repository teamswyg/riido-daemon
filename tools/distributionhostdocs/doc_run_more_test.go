package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaybeWriteAndCheckDocs(t *testing.T) {
	dir := t.TempDir()
	docs := map[string]string{"docs/out.md": "fresh"}
	if err := maybeWriteDocs(options{Repo: dir}, docs); err != nil {
		t.Fatal(err)
	}
	if problems := checkDocs(options{Repo: dir, CheckDoc: true}, docs); len(problems) != 1 {
		t.Fatalf("expected missing doc problem, got %#v", problems)
	}
	if err := maybeWriteDocs(options{Repo: dir, WriteDoc: true}, docs); err != nil {
		t.Fatal(err)
	}
	mustDistributionHostWrite(t, filepath.Join(dir, "docs/out.md"), "stale")
	assertDistributionHostProblem(t, checkDocs(options{Repo: dir, CheckDoc: true}, docs), "generated doc drift")
}

func TestRunWritesFailedEvidenceForInvalidManifest(t *testing.T) {
	dir := t.TempDir()
	mustDistributionHostWrite(t, filepath.Join(dir, "manifest.json"), `{
  "schema_version": "bad",
  "id": "x",
  "generated_doc": "doc.md",
  "workflow": "missing.yml",
  "evidence_artifact": "artifact",
  "parts": [],
  "related_pages": [],
  "fragments": [],
  "source_checks": [],
  "assertions": []
}`)
	out := filepath.Join(dir, "out", "evidence.json")
	err := run(options{Repo: dir, Manifest: "manifest.json", EvidenceOut: out})
	if err == nil || !strings.Contains(err.Error(), "unexpected schema_version") {
		t.Fatalf("expected validation error, got %v", err)
	}
	data, readErr := os.ReadFile(out)
	if readErr != nil || !strings.Contains(string(data), `"status": "failed"`) {
		t.Fatalf("evidence=%q err=%v", data, readErr)
	}
}

func TestSourceCheckAndPathsFailures(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "source.txt")
	mustDistributionHostWrite(t, target, "anchor")
	if !filepath.IsAbs(repoPath(dir, target)) {
		t.Fatal("absolute repo path should stay absolute")
	}
	check := sourceCheck{Name: "n", File: "source.txt", Contains: "anchor"}
	if passed, err := sourceCheckPasses(dir, check); !passed || err != nil {
		t.Fatalf("passed=%v err=%v", passed, err)
	}
	if _, err := sourceCheckPasses(dir, sourceCheck{Name: "n"}); err == nil {
		t.Fatal("expected incomplete source check to fail")
	}
	assertDistributionHostProblem(t, mustExist(dir, "missing.yml"), "missing artifact")
}

func mustDistributionHostWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := writeText(path, body); err != nil {
		t.Fatal(err)
	}
}
