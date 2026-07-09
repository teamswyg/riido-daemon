package main

import (
	"strings"
	"testing"
)

func TestValidateBlockReportsInvalidShapes(t *testing.T) {
	blocks := []block{
		{Kind: "heading"},
		{Kind: "paragraph"},
		{Kind: "bullets"},
		{Kind: "links", Links: []link{{Title: "T"}}},
		{Kind: "table"},
		{Kind: "table", Columns: []string{"a"}, Rows: [][]string{{"x", "y"}}},
		{Kind: "code", Language: "go"},
		{Kind: "unknown"},
	}
	var problems []string
	for _, block := range blocks {
		problems = append(problems, validateBlock("page", block)...)
	}
	for _, want := range []string{
		"empty text block",
		"empty list",
		"links require title and path",
		"empty table",
		"table row width mismatch",
		"empty code block",
		"invalid block kind unknown",
	} {
		assertReadmeProblem(t, problems, want)
	}
}

func TestValidatePageReportsSchemaRequiredAndDuplicateIssues(t *testing.T) {
	seen := map[string]bool{"same": true}
	problems := validatePage(page{ID: "same", SchemaVersion: "bad"}, seen)
	assertReadmeProblem(t, problems, "unexpected page schema_version")
	assertReadmeProblem(t, problems, "page id, title, generated_doc, and blocks are required")
	assertReadmeProblem(t, problems, "duplicate page id same")
}

func TestValidateManifestReportsMissingShape(t *testing.T) {
	problems, _ := validateManifest(t.TempDir(), manifest{})
	assertReadmeProblem(t, problems, "unexpected schema_version")
	assertReadmeProblem(t, problems, "id, workflow, and evidence_artifact are required")
	assertReadmeProblem(t, problems, "assertions are required")
	problems, _ = validateManifest(t.TempDir(), manifest{Workflow: "missing.yml"})
	assertReadmeProblem(t, problems, "missing artifact")
}

func assertReadmeProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, p := range problems {
		if strings.Contains(p, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}
