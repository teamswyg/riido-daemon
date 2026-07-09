package main

import (
	"strings"
	"testing"
)

func TestValidateBlocksReportsInvalidShapes(t *testing.T) {
	detail := detailDoc{ID: "detail", Blocks: []block{
		{Kind: "paragraph"},
		{Kind: "bullets"},
		{Kind: "ordered"},
		{Kind: "table"},
		{Kind: "table", Columns: []string{"a"}, Rows: [][]string{{"x", "y"}}},
		{Kind: "code", Language: "go"},
		{Kind: "unknown"},
	}}
	problems := validateBlocks(detail)
	for _, want := range []string{
		"empty paragraph",
		"empty list",
		"empty table",
		"table row width mismatch",
		"empty code block",
		"invalid block kind unknown",
	} {
		assertSecurityRedactionProblem(t, problems, want)
	}
}

func TestValidateDetailsReportsSchemaRequiredAndDuplicateIssues(t *testing.T) {
	m := manifest{Details: []detailDoc{
		{ID: "same", SchemaVersion: "bad"},
		{ID: "same", SchemaVersion: detailSchema, Title: "T", GeneratedDoc: "d.md", Blocks: []block{{Kind: "paragraph", Text: "body"}}},
	}}
	problems := validateDetails(m)
	assertSecurityRedactionProblem(t, problems, "unexpected detail schema_version")
	assertSecurityRedactionProblem(t, problems, "detail id, title, generated_doc, and blocks are required")
	assertSecurityRedactionProblem(t, problems, "duplicate detail id same")
}

func assertSecurityRedactionProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}
