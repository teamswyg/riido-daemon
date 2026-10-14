package main

import (
	"fmt"
	"strings"
)

func renderInputs(b *strings.Builder, rows []Input) {
	b.WriteString("## Input Surfaces\n\n")
	for _, row := range rows {
		fmt.Fprintf(b, "- `%s`: %s Evidence: %s.\n", row.Name, row.Summary, checks(row.SourceChecks))
	}
	b.WriteByte('\n')
}

func renderExample(b *strings.Builder) {
	b.WriteString("## Local Queue Example\n\n")
	b.WriteString("```json\n")
	b.WriteString("{\n")
	b.WriteString("  \"id\": \"task-1\",\n")
	b.WriteString("  \"provider\": \"cursor\",\n")
	b.WriteString("  \"prompt\": \"inspect this repo\",\n")
	b.WriteString("  \"required_surfaces\": [\"structured-event-stream\", \"worktree\"],\n")
	b.WriteString("  \"allow_experimental_runtime\": true,\n")
	b.WriteString("  \"metadata\": {\"workspace_id\": \"ws-1\"}\n")
	b.WriteString("}\n")
	b.WriteString("```\n\n")
}

func renderAssertions(b *strings.Builder, assertions []string) {
	b.WriteString("## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- %s\n", assertion)
	}
}
