package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesVerifiedEvidence(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	workflowPath := filepath.Join(dir, ".github", "workflows", "example.yml")
	mustWrite(t, manifestPath, `{"schema_version":"riido-daemon-package-workflow-evidence.v1","id":"m","loop_source":"loops/m.json","workflows":[{"id":"example","workflow":"`+filepath.ToSlash(workflowPath)+`","evidence_artifact":"example","required_fragments":["go test ./example"]}]}`)
	mustWrite(t, workflowPath, "run: go test ./example\n")
	out := filepath.Join(dir, "out", "evidence.json")
	if err := run([]string{"-manifest", manifestPath, "-workflow", workflowPath, "-evidence-out", out}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
