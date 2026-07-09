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
		{Kind: "code", Language: "go"},
		{Kind: "unknown"},
	}}
	problems := validateBlocks(detail)
	for _, want := range []string{
		"empty heading",
		"empty paragraph",
		"empty list",
		"empty table",
		"table row width mismatch",
		"empty code block",
		"invalid block kind unknown",
	} {
		assertProviderAdapterACLProblem(t, problems, want)
	}
}

func TestValidateDetailsReportsSchemaRequiredAndDuplicateIssues(t *testing.T) {
	m := manifest{Details: []detailDoc{
		{ID: "same", SchemaVersion: "bad"},
		{ID: "same", SchemaVersion: detailSchema, Title: "T", GeneratedDoc: "d.md", Blocks: []block{{Kind: "paragraph", Text: "body"}}},
	}}
	problems := validateDetails(m)
	assertProviderAdapterACLProblem(t, problems, "unexpected detail schema_version")
	assertProviderAdapterACLProblem(t, problems, "detail id, title, generated_doc, and blocks are required")
	assertProviderAdapterACLProblem(t, problems, "duplicate detail id same")
}

func assertProviderAdapterACLProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}
