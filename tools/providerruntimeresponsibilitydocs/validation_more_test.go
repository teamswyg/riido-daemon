package main

import (
	"strings"
	"testing"
)

func TestValidateBlocksReportsAllInvalidShapes(t *testing.T) {
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
		assertResponsibilityProblem(t, problems, want)
	}
}

func TestValidateDetailsReportsSchemaRequiredAndDuplicateIssues(t *testing.T) {
	m := manifest{Details: []detailDoc{
		{ID: "same", SchemaVersion: "bad"},
		{ID: "same", SchemaVersion: detailSchema, Title: "T", GeneratedDoc: "d.md", Blocks: []block{{Kind: "paragraph", Text: "body"}}},
	}}
	problems := validateDetails(m)
	assertResponsibilityProblem(t, problems, "unexpected detail schema_version")
	assertResponsibilityProblem(t, problems, "detail id, title, generated_doc, and blocks are required")
	assertResponsibilityProblem(t, problems, "duplicate detail id same")
}

func TestRenderBlockCoversOrderedAndCode(t *testing.T) {
	var b strings.Builder
	renderBlock(&b, block{Kind: "ordered", Items: []string{"one", "two"}})
	renderBlock(&b, block{Kind: "code", Language: "go", Code: "fmt.Println()"})
	out := b.String()
	if !strings.Contains(out, "1. one") || !strings.Contains(out, "```go\nfmt.Println()") {
		t.Fatalf("rendered block = %q", out)
	}
}

func assertResponsibilityProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, p := range problems {
		if strings.Contains(p, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}
