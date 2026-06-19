package main

import (
	"os"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "docs", "a.md"), "# A\n")
	mustWrite(t, filepath.Join(dir, "docs", "README.md"), "")
	mustWrite(t, filepath.Join(dir, "docs", "readme", "document-map.md"), "")
	manifestPath := filepath.Join(dir, "docs", "readme", "document-map.riido.json")
	data := `{"schema_version":"riido-doc-map.v1","id":"test","title":"Test","generated_docs":{"readme":"docs/README.md","document_map":"docs/readme/document-map.md"},"evidence_artifact":"artifact","intro":"intro","read_order":[{"doc":"docs/a.md","description":"A"}],"decisions":[{"topic":"A","docs":["docs/a.md"]}],"repos":[{"repo":"r","responsibility":"x"}],"rules":["rule"]}`
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
