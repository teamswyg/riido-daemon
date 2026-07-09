package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssignmentFSMEvidenceAndProblemHelpers(t *testing.T) {
	manifest := Manifest{
		ID:                 "assignment",
		SchemaVersion:      "riido-assignment-fsm.v1",
		GeneratedDoc:       "docs/fsm.md",
		Workflow:           ".github/workflows/fsm.yml",
		SourcePackage:      "assignment",
		ForbiddenDocTokens: []string{"blocked"},
	}
	fsm := FSMSnapshot{Name: "assignment", States: []string{"queued"}}
	problems := []problem{{Message: "generated doc drift"}}
	sources := []SourceCheckEvidence{{Name: "contract", File: "state.go", OK: true}}
	ev := buildEvidence(manifest, fsm, problems, sources, false)
	if ev.ID != "assignment" || ev.FSM.Name != "assignment" || ev.ForbiddenCheck.OK {
		t.Fatalf("unexpected evidence=%+v", ev)
	}
	if !strings.Contains(problemError(problems).Error(), "generated doc drift") {
		t.Fatal("problem error should include problem messages")
	}
}

func TestAssignmentFSMIOWritesAndRejectsNonSemanticJSON(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "nested", "doc.md")
	if err := writeText(textPath, "doc"); err == nil {
		t.Fatal("writeText should require existing parent directory")
	}
	if err := os.MkdirAll(filepath.Dir(textPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeText(textPath, "doc"); err != nil {
		t.Fatal(err)
	}
	jsonPath := filepath.Join(dir, "evidence.json")
	if err := writeJSON(jsonPath, map[string]string{"id": "assignment"}); err != nil {
		t.Fatal(err)
	}
	trailing := filepath.Join(dir, "trailing.json")
	if err := os.WriteFile(trailing, []byte(`{} {}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadManifest("", trailing); err == nil {
		t.Fatal("trailing JSON values must fail")
	}
	unknown := filepath.Join(dir, "unknown.json")
	if err := os.WriteFile(unknown, []byte(`{"unknown":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadManifest("", unknown); err == nil {
		t.Fatal("unknown manifest fields must fail")
	}
}
