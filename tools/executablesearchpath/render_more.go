package main

import (
	"fmt"
	"strings"
)

func renderRules(b *strings.Builder, rows []Rule) {
	b.WriteString("## Rules\n\n")
	b.WriteString("| Rule | Meaning | Behavior | Evidence |\n| --- | --- | --- | --- |\n")
	for _, row := range rows {
		fmt.Fprintf(b, "| %s | %s | `%s` | %s |\n",
			row.Name, row.Summary, behaviorLabel(row.Behavior), checks(row.SourceChecks))
	}
	b.WriteByte('\n')
}

func behaviorLabel(value string) string {
	if value == "" {
		return "source-check"
	}
	return value
}

func renderAssertions(b *strings.Builder, assertions []string) {
	b.WriteString("## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- %s\n", assertion)
	}
}
