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
		assertBoundaryProblem(t, problems, want)
	}
}

func TestValidateDetailsReportsSchemaRequiredAndDuplicateIssues(t *testing.T) {
	m := manifest{Details: []detailDoc{
		{ID: "same", SchemaVersion: "bad"},
		{ID: "same", SchemaVersion: detailSchema, Title: "T", GeneratedDoc: "d.md", Blocks: []block{{Kind: "paragraph", Text: "body"}}},
	}}
	problems := validateDetails(m)
	assertBoundaryProblem(t, problems, "unexpected detail schema_version")
	assertBoundaryProblem(t, problems, "detail id, title, generated_doc, and blocks are required")
	assertBoundaryProblem(t, problems, "duplicate detail id same")
}

func TestRenderBlockCoversOrderedList(t *testing.T) {
	var b strings.Builder
	renderBlock(&b, block{Kind: "ordered", Items: []string{"one", "two"}})
	out := b.String()
	if !strings.Contains(out, "1. one") || !strings.Contains(out, "2. two") {
		t.Fatalf("rendered block = %q", out)
	}
}

func assertBoundaryProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, p := range problems {
		if strings.Contains(p, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}
