package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoopEvidenceLoadsLoopFiles(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "loop.riido.json")
	docPath := filepath.Join(dir, "loop.md")
	loopPath := filepath.Join(dir, "loops", "x.riido.json")
	writeLoopFixture(t, loopPath)
	data := `{"schema_version":"riido-loop-evidence.v1","id":"ok","title":"OK","generated_doc":"loop.md","required_phases":["observe","hypothesis","execute","evaluate","retrospective"],"loop_files":["loops/x.riido.json"],"open_gaps":[]}`
	if err := os.WriteFile(manifestPath, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	rendered := renderMarkdown(mustLoadExpanded(t, dir, manifestPath))
	if err := os.WriteFile(docPath, []byte(rendered), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(dir, manifestPath, docPath, false, true); err != nil {
		t.Fatal(err)
	}
}

func mustLoadExpanded(t *testing.T, root, path string) manifest {
	t.Helper()
	loaded, err := loadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err = expandLoopFiles(root, loaded)
	if err != nil {
		t.Fatal(err)
	}
	return loaded
}
