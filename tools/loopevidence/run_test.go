package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoopEvidenceCurrentManifest(t *testing.T) {
	err := run("../..", "docs/30-architecture/loop-engineering.riido.json", "", false, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoopEvidenceGeneratedDocCurrent(t *testing.T) {
	err := run("../..", "docs/30-architecture/loop-engineering.riido.json", "", false, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoopEvidenceRejectsMissingPhase(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "loop.riido.json")
	docPath := filepath.Join(dir, "loop.md")
	data := `{
	  "schema_version":"riido-loop-evidence.v1",
	  "id":"bad",
	  "title":"Bad",
	  "generated_doc":"loop.md",
	  "required_phases":["observe","hypothesis","execute","evaluate","retrospective"],
	  "loops":[{"id":"x","owner":"test"}],
	  "open_gaps":[{"id":"gap","owner":"test","current_handling":"x","required_next_artifact":"y"}]
	}`
	if err := os.WriteFile(manifestPath, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(docPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	err := run(dir, manifestPath, docPath, false, false)
	if err == nil || !strings.Contains(err.Error(), "summary is required") {
		t.Fatalf("expected missing phase error, got %v", err)
	}
}
