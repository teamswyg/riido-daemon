package main

import (
	"strings"
	"testing"
)

func TestValidateBlocksReportsInvalidShapes(t *testing.T) {
	detail := detailDoc{ID: "detail", Blocks: []block{
		{Kind: "heading"},
		{Kind: "paragraph"},
		{Kind: "bullets"},
		{Kind: "ordered"},
		{Kind: "table"},
		{Kind: "table", Columns: []string{"a"}, Rows: [][]string{{"x", "y"}}},
		{Kind: "unknown"},
	}}
	problems := validateBlocks(detail)
	for _, want := range []string{
		"empty heading",
		"empty paragraph",
		"empty list",
		"empty table",
		"table row width mismatch",
		"invalid block kind unknown",
	} {
		assertWorkspaceNativeProblem(t, problems, want)
	}
}

func TestValidateDetailsReportsSchemaRequiredAndDuplicateIssues(t *testing.T) {
	m := manifest{Details: []detailDoc{
		{ID: "same", SchemaVersion: "bad"},
		{ID: "same", SchemaVersion: detailSchema, Title: "T", GeneratedDoc: "d.md", Blocks: []block{{Kind: "paragraph", Text: "body"}}},
	}}
	problems := validateDetails(m)
	assertWorkspaceNativeProblem(t, problems, "unexpected detail schema_version")
	assertWorkspaceNativeProblem(t, problems, "detail id, title, generated_doc, and blocks are required")
	assertWorkspaceNativeProblem(t, problems, "duplicate detail id same")
}

func assertWorkspaceNativeProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, p := range problems {
		if strings.Contains(p, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}
