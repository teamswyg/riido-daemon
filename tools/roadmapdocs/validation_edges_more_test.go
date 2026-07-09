package main

import (
	"path/filepath"
	"testing"
)

func TestValidateManifestReportsRequiredAndWorkflowProblems(t *testing.T) {
	m := manifest{
		SchemaVersion:    "bad",
		Workflow:         "missing.yml",
		EvidenceArtifact: "",
		Questions:        nil,
		Assertions:       nil,
	}
	problems, checks := validateManifest(t.TempDir(), m)
	if len(checks) != 0 {
		t.Fatalf("unexpected source checks %#v", checks)
	}
	for _, want := range []string{
		"unexpected schema_version",
		"id, title, and generated_doc are required",
		"workflow and evidence_artifact are required",
		"questions and assertions are required",
		"missing artifact \"missing.yml\"",
	} {
		if !containsProblem(problems, want) {
			t.Fatalf("missing %q in %#v", want, problems)
		}
	}
}

func TestSourceCheckPassesReportsRequiredMissingAndSuccess(t *testing.T) {
	dir := t.TempDir()
	if ok, err := sourceCheckPasses(dir, sourceCheck{}); err == nil || ok {
		t.Fatalf("expected required field failure ok=%v err=%v", ok, err)
	}
	if ok, err := sourceCheckPasses(dir, sourceCheck{Name: "missing", File: "no.md", Contains: "x"}); err == nil || ok {
		t.Fatalf("expected missing file failure ok=%v err=%v", ok, err)
	}
	mustRoadmapWrite(t, filepath.Join(dir, "doc.md"), "anchor")
	if ok, err := sourceCheckPasses(dir, sourceCheck{Name: "ok", File: "doc.md", Contains: "anchor"}); err != nil || !ok {
		t.Fatalf("expected source check pass ok=%v err=%v", ok, err)
	}
}
