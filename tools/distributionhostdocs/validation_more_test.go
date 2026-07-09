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
		assertDistributionHostProblem(t, validateBlock("page", tc.block), tc.want)
	}
}

func TestValidatePagesReportsSchemaRequiredAndDuplicateIssues(t *testing.T) {
	pages := []page{
		{ID: "same", SchemaVersion: "bad"},
		{ID: "same", SchemaVersion: pageSchema, Title: "T", GeneratedDoc: "p.md", Blocks: []block{{Kind: "paragraph", Text: "body"}}},
		{ID: "third", SchemaVersion: pageSchema, Title: "T", GeneratedDoc: "p.md", Blocks: []block{{Kind: "paragraph", Text: "body"}}},
		{ID: "fourth", SchemaVersion: pageSchema, Title: "T", GeneratedDoc: "p.md", Blocks: []block{{Kind: "paragraph", Text: "body"}}},
	}
	problems := validatePages(pages)
	assertDistributionHostProblem(t, problems, "unexpected page schema_version")
	assertDistributionHostProblem(t, problems, "page id, title, generated_doc, and blocks are required")
	assertDistributionHostProblem(t, problems, "duplicate page id same")
}

func TestValidatePagesRequiresFourPages(t *testing.T) {
	assertDistributionHostProblem(t, validatePages(nil), "four generated distribution host pages are required")
}

func assertDistributionHostProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}
