package main

import (
	"path/filepath"
	"testing"
)

func TestValidateSourceChecksAndBlocks(t *testing.T) {
	dir := t.TempDir()
	if err := writeText(filepath.Join(dir, "lock.go"), "anchor"); err != nil {
		t.Fatal(err)
	}
	checks, problems := validateSourceChecks(dir, []sourceCheck{
		{Name: "", File: "lock.go", Contains: "anchor"},
		{Name: "missing", File: "lock.go", Contains: "absent"},
		{Name: "ok", File: "lock.go", Contains: "anchor"},
	})
	for _, want := range []string{
		"source check requires name, file, and contains",
		"source check \"missing\" missing anchor",
	} {
		assertLockingProblem(t, problems, want)
	}
	if len(checks) != 3 || !checks[2].Passed {
		t.Fatalf("unexpected source checks %#v", checks)
	}
	detail := detailDoc{ID: "detail", Blocks: []block{
		{Kind: "paragraph"},
		{Kind: "bullets"},
		{Kind: "table", Columns: []string{"a"}, Rows: [][]string{{"a", "b"}}},
		{Kind: "code"},
		{Kind: "unknown"},
	}}
	for _, want := range []string{
		"detail has empty paragraph",
		"detail has empty list",
		"detail table row width mismatch",
		"detail has empty code block",
		"detail has invalid block kind unknown",
	} {
		assertLockingProblem(t, validateBlocks(detail), want)
	}
}

func TestValidateManifestAndEvidenceStatus(t *testing.T) {
	dir := t.TempDir()
	m := manifest{
		SchemaVersion: "bad",
		DetailPages:   []detailLink{{ID: "detail"}},
		Details: []detailDoc{{
			SchemaVersion: "bad", ID: "detail", Blocks: []block{{Kind: "paragraph"}},
		}},
		Workflow:         "missing.yml",
		EvidenceArtifact: "evidence.json",
	}
	problems, checks := validateManifest(dir, m)
	for _, want := range []string{
		"unexpected schema_version",
		"id, generated_doc, workflow, and evidence_artifact are required",
		"unexpected detail schema_version: detail",
		"detail id, title, generated_doc, and blocks are required",
		"missing artifact \"missing.yml\"",
	} {
		assertLockingProblem(t, problems, want)
	}
	ev := buildEvidence(m, checks, problems[:1])
	if ev.Status != "failed" || ev.DetailCount != 1 || ev.EvidenceArtifact != "evidence.json" {
		t.Fatalf("unexpected evidence %#v", ev)
	}
}
