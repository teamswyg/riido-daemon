package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSourceEventKindsAndManifestKinds(t *testing.T) {
	dir := t.TempDir()
	sourceDir := filepath.Join(dir, "internal", "agentbridge")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `package agentbridge
type EventKind string
const (
	EventAlpha EventKind = "alpha"
	EventBeta EventKind = "beta"
)
const NotEvent = "ignored"
`
	if err := os.WriteFile(filepath.Join(sourceDir, "event_kind.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	kinds, err := sourceEventKinds(dir)
	if err != nil || len(kinds) != 2 {
		t.Fatalf("kinds=%#v err=%v", kinds, err)
	}
	manifestKinds, problems := manifestKindMap(Manifest{
		SemanticActivity:    []string{"alpha", "dup", ""},
		NonSemanticActivity: []string{"dup", "beta"},
	})
	for _, want := range []string{"empty event kind", "duplicate event kind: dup"} {
		assertSemanticProblem(t, problems, want)
	}
	if len(manifestKinds) != 3 || !manifestKinds["alpha"] || manifestKinds["beta"] {
		t.Fatalf("unexpected manifest kinds %#v", manifestKinds)
	}
}

func TestValidateAndEvidenceReportDrift(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "internal", "agentbridge"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "internal", "agentbridge", "event_kind.go"), []byte("package agentbridge\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := Manifest{
		SchemaVersion:       "bad",
		ID:                  "semantic",
		Title:               "Semantic",
		GeneratedDoc:        "doc.md",
		Workflow:            "workflow.yml",
		EvidenceArtifact:    "evidence.json",
		SemanticActivity:    []string{"unknown"},
		NonSemanticActivity: []string{"unknown"},
		Assertions:          []string{"events classify activity"},
	}
	problems := validate(dir, m)
	for _, want := range []string{
		"schema_version must be riido-semantic-event-activity.v1",
		"no EventKind source declarations found",
		"duplicate event kind: unknown",
		"manifest declares unknown event kind: unknown",
	} {
		assertSemanticProblem(t, problems, want)
	}
	ev := buildEvidence(m, problems[:1])
	if ev.Status != "failed" || ev.ID != "semantic" || len(ev.Assertions) != 1 {
		t.Fatalf("unexpected evidence %#v", ev)
	}
}
