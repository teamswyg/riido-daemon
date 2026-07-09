package main

import (
	"strings"
	"testing"
)

func TestRenderPageTableCodeAndLinks(t *testing.T) {
	p := page{
		SchemaVersion: pageSchema,
		ID:            "table-page",
		Title:         "Table Page",
		GeneratedDoc:  "table.md",
		Blocks: []block{
			{Kind: "table", Columns: []string{"Field", "Meaning"}, Rows: [][]string{{"runtime", "provider"}}},
			{Kind: "code", Language: "json", Code: `{"ok":true}`},
			{Kind: "links", Links: []link{{Title: "runtime", Path: "runtime.md"}}},
		},
	}
	body := renderPage(p)
	for _, want := range []string{"| Field | Meaning |", "| runtime | provider |", "```json", "[runtime](runtime.md)"} {
		if !strings.Contains(body, want) {
			t.Fatalf("rendered page missing %q:\n%s", want, body)
		}
	}
}

func TestValidateBlockProblems(t *testing.T) {
	cases := []struct {
		name  string
		block block
		want  string
	}{
		{name: "empty text", block: block{Kind: "heading"}, want: "empty text"},
		{name: "empty list", block: block{Kind: "bullets"}, want: "empty list"},
		{name: "bad link", block: block{Kind: "links", Links: []link{{Title: "x"}}}, want: "links require"},
		{name: "empty table", block: block{Kind: "table"}, want: "empty table"},
		{name: "row mismatch", block: block{Kind: "table", Columns: []string{"a"}, Rows: [][]string{{"a", "b"}}}, want: "row width"},
		{name: "empty code", block: block{Kind: "code", Language: "go"}, want: "empty code"},
		{name: "invalid", block: block{Kind: "unknown"}, want: "invalid block"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			problems := validateBlock("page", tc.block)
			if len(problems) == 0 || !strings.Contains(problems[0], tc.want) {
				t.Fatalf("problems = %+v, want %q", problems, tc.want)
			}
		})
	}
}

func TestValidatePageDuplicateAndSchema(t *testing.T) {
	seen := map[string]bool{"dup": true}
	p := page{SchemaVersion: "bad", ID: "dup"}
	problems := validatePage(p, seen)
	for _, want := range []string{"unexpected page schema_version", "required", "duplicate page id"} {
		if !hasProblem(problems, want) {
			t.Fatalf("problems = %+v, missing %q", problems, want)
		}
	}
}

func hasProblem(problems []string, want string) bool {
	for _, problem := range problems {
		if strings.Contains(problem, want) {
			return true
		}
	}
	return false
}
