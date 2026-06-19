package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoopEvidenceAllowsNoOpenGaps(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "loop.riido.json")
	docPath := filepath.Join(dir, "loop.md")
	data := `{
	  "schema_version":"riido-loop-evidence.v1",
	  "id":"ok",
	  "title":"OK",
	  "generated_doc":"loop.md",
	  "required_phases":["observe","hypothesis","execute","evaluate","retrospective"],
	  "loops":[{"id":"x","owner":"test","observation":{"summary":"o"},"hypothesis":{"summary":"h"},"execution":{"summary":"e"},"evaluation":{"summary":"v"},"retrospective":{"summary":"r"},"evidence":[{"kind":"command","ref":"echo ok","proves":"ok"}]}],
	  "open_gaps":[]
	}`
	if err := os.WriteFile(manifestPath, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(docPath, []byte(renderMarkdown(mustLoadLoop(t, manifestPath))), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(options{Repo: dir, Manifest: manifestPath, Doc: docPath, Check: true}); err != nil {
		t.Fatal(err)
	}
}

func mustLoadLoop(t *testing.T, path string) manifest {
	t.Helper()
	out, err := loadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
