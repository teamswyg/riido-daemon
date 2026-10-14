package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, dir string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, "src.go"), "needle")
	mustWrite(t, filepath.Join(dir, "validation.md"), "")
	manifest := `{"schema_version":"riido-validation-evidence.v1","id":"x","title":"X","generated_doc":"validation.md","workflow":".github/workflows/x.yml","evidence_artifact":"x","purpose":"p","facts":[{"name":"f","summary":"s","source_checks":["src"]}],"boundaries":[{"name":"b","owner":"o","summary":"s"}],"absent_surfaces":[{"name":"a","scope":["src.go"],"tokens":["missing"],"reason":"r"}],"source_checks":[{"name":"src","file":"src.go","contains":"needle"}],"assertions":["a"]}`
	mustWrite(t, filepath.Join(dir, "manifest.json"), manifest)
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
