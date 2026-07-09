package main

import (
	"path/filepath"
	"testing"
)

func TestValidateSourceChecksAndPageBlocks(t *testing.T) {
	dir := t.TempDir()
	if err := writeText(filepath.Join(dir, "migration.go"), "anchor"); err != nil {
		t.Fatal(err)
	}
	checks, problems := validateSourceChecks(dir, []sourceCheck{
		{Name: "", File: "migration.go", Contains: "anchor"},
		{Name: "missing", File: "migration.go", Contains: "absent"},
		{Name: "ok", File: "migration.go", Contains: "anchor"},
	})
	for _, want := range []string{
		"source check requires name, file, and contains",
		"source check \"missing\" missing anchor",
	} {
		assertMigrationProblem(t, problems, want)
	}
	if len(checks) != 3 || !checks[2].Passed {
		t.Fatalf("unexpected source checks %#v", checks)
	}
	page := page{ID: "page", Blocks: []block{
		{Kind: "heading"},
		{Kind: "bullets"},
		{Kind: "links", Links: []link{{Title: "Missing path"}}},
		{Kind: "table", Columns: []string{"a"}, Rows: [][]string{{"a", "b"}}},
		{Kind: "code"},
		{Kind: "unknown"},
	}}
	for _, want := range []string{
		"page has empty text block",
		"page has empty list",
		"page links require title and path",
		"page table row width mismatch",
		"page has empty code block",
		"page has invalid block kind unknown",
	} {
		assertMigrationProblem(t, validatePage(page, map[string]bool{}), want)
	}
}

func TestValidateManifestAndEvidenceStatus(t *testing.T) {
	dir := t.TempDir()
	m := manifest{
		SchemaVersion:     "bad",
		Workflow:          "missing.yml",
		EvidenceArtifact:  "evidence.json",
		ExpectedPageCount: 2,
		Pages: []page{{
			SchemaVersion: "bad", ID: "page", Blocks: []block{{Kind: "paragraph"}},
		}},
	}
	problems, checks := validateManifest(dir, m)
	for _, want := range []string{
		"unexpected schema_version",
		"id, workflow, and evidence_artifact are required",
		"assertions are required",
		"expected 2 migration pages, got 1",
		"missing artifact \"missing.yml\"",
	} {
		assertMigrationProblem(t, problems, want)
	}
	ev := buildEvidence(m, checks, problems[:1])
	if ev.Status != "failed" || ev.EvidenceArtifact != "evidence.json" {
		t.Fatalf("unexpected evidence %#v", ev)
	}
}
