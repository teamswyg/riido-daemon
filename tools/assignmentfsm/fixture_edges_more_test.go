package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFSMFixture(t *testing.T, repo, source string) {
	t.Helper()
	manifest := `{"schema_version":"riido-assignment-fsm.v1","id":"fsm","title":"FSM","generated_doc":"docs/fsm.md","workflow":"w","evidence_artifact":"e","source_package":"pkg","evidence_manifest":"m","purpose":"p","source_checks":[{"name":"src","file":"src.go","contains":"needle"}],"forbidden_doc_tokens":["legacy"],"related_sections":[]}`
	writeFile(t, filepath.Join(repo, "manifest.json"), manifest)
	writeFile(t, filepath.Join(repo, "src.go"), source)
	if err := os.MkdirAll(filepath.Join(repo, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
