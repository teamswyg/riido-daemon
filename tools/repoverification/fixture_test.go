package main

import (
	"os"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "workflow.yml"), "name: test\n")
	mustWrite(t, filepath.Join(dir, "doc.md"), "")
	manifestPath := filepath.Join(dir, "manifest.json")
	data := `{"schema_version":"riido-repo-verification.v1","id":"test","title":"Test","generated_doc":"doc.md","workflow":"workflow.yml","evidence_artifact":"artifact","commands":[{"id":"a","description":"A","argv":["echo","a"]},{"id":"b","description":"B","argv":["echo","b"]}],"assertions":["ok"]}`
	mustWrite(t, manifestPath, data)
	return dir, manifestPath
}

func mustWrite(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}
