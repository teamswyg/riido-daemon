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
	mustRuntimeSchedulingWrite(t, filepath.Join(dir, "docs/out.md"), "stale")
	problems := checkDocs(options{Repo: dir, CheckDoc: true}, docs)
	assertRuntimeSchedulingProblem(t, problems, "generated doc drift")
}

func TestRunWritesFailedEvidenceForInvalidManifest(t *testing.T) {
	dir := t.TempDir()
	mustRuntimeSchedulingWrite(t, filepath.Join(dir, "core.json"), `{"schema_version":"bad","id":"core","title":"Core","generated_doc":"core.md","context":"ctx","responsibilities":["r"],"non_responsibilities":["n"],"invariants":[]}`)
	mustRuntimeSchedulingWrite(t, filepath.Join(dir, "manifest.json"), `{
  "schema_version": "bad",
  "id": "x",
  "generated_doc": "doc.md",
  "workflow": "missing.yml",
  "evidence_artifact": "artifact",
  "parts": [],
  "invariants_index": {},
  "core_fragment": "core.json",
  "evidence_pages": [],
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

func mustRuntimeSchedulingWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
