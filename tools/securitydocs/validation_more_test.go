package main

import (
	"strings"
	"testing"
)

func TestValidateBlockReportsInvalidShapes(t *testing.T) {
	cases := []struct {
		block block
		want  string
	}{
		{block{Kind: "heading"}, "empty text block"},
		{block{Kind: "paragraph"}, "empty text block"},
		{block{Kind: "bullets"}, "empty list"},
		{block{Kind: "links"}, "empty links"},
		{block{Kind: "links", Links: []link{{Title: "T"}}}, "links require title and path"},
		{block{Kind: "table"}, "empty table"},
		{block{Kind: "table", Columns: []string{"a"}, Rows: [][]string{{"x", "y"}}}, "table row width mismatch"},
		{block{Kind: "code", Language: "go"}, "empty code block"},
		{block{Kind: "unknown"}, "invalid block kind unknown"},
	}
	for _, tc := range cases {
		assertSecurityDocsProblem(t, validateBlock("page", tc.block), tc.want)
	}
}

func TestValidatePagesReportsSchemaRequiredAndDuplicateIssues(t *testing.T) {
	pages := []page{
		{ID: "same", SchemaVersion: "bad"},
		{ID: "same", SchemaVersion: pageSchema, Title: "T", GeneratedDoc: "p.md", Blocks: []block{{Kind: "paragraph", Text: "body"}}},
		{ID: "third", SchemaVersion: pageSchema, Title: "T", GeneratedDoc: "p.md", Blocks: []block{{Kind: "paragraph", Text: "body"}}},
	}
	problems := validatePages(pages)
	assertSecurityDocsProblem(t, problems, "unexpected page schema_version")
	assertSecurityDocsProblem(t, problems, "page id, title, generated_doc, and blocks are required")
	assertSecurityDocsProblem(t, problems, "duplicate page id same")
}

func TestValidatePagesRequiresThreePages(t *testing.T) {
	assertSecurityDocsProblem(t, validatePages(nil), "three security wrapper pages are required")
}

func assertSecurityDocsProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}
