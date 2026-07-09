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
	mustContextMapWrite(t, filepath.Join(dir, "docs/out.md"), "stale")
	assertContextMapProblem(t, checkDocs(options{Repo: dir, CheckDoc: true}, docs), "generated doc drift")
}

func TestRunWritesFailedEvidenceForInvalidLoadedManifest(t *testing.T) {
	dir := t.TempDir()
	writeContextMapManifestFixture(t, dir)
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

func writeContextMapManifestFixture(t *testing.T, dir string) {
	t.Helper()
	mustContextMapWrite(t, filepath.Join(dir, "manifest.json"), `{
  "schema_version": "bad",
  "id": "context",
  "generated_doc": "context.md",
  "workflow": "workflow.yml",
  "evidence_artifact": "evidence.json",
  "focused_sections": [],
  "fragments": {
    "acl_locations": "acl.json",
    "dependency_direction": "dependency.json",
    "figma_daemon_boundaries": "figma.json",
    "figma_onboarding_boundaries": "onboarding.json",
    "split_repo_ownership": "split.json",
    "change_procedure": "change.json"
  },
  "contexts": []
}`)
	mustContextMapWrite(t, filepath.Join(dir, "acl.json"), `{"schema_version":"x","rows":[]}`)
	mustContextMapWrite(t, filepath.Join(dir, "dependency.json"), `{"schema_version":"x","diagram":[]}`)
	mustContextMapWrite(t, filepath.Join(dir, "figma.json"), `{"schema_version":"x","sections":[]}`)
	mustContextMapWrite(t, filepath.Join(dir, "onboarding.json"), `{"schema_version":"x","sections":[]}`)
	mustContextMapWrite(t, filepath.Join(dir, "split.json"), `{"schema_version":"x","rules":[]}`)
	mustContextMapWrite(t, filepath.Join(dir, "change.json"), `{"schema_version":"x","same_pr_updates":[]}`)
}

func mustContextMapWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := writeText(path, body); err != nil {
		t.Fatal(err)
	}
}
