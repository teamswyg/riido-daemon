package main

import (
	"fmt"
	"strings"
)

func renderPassthroughs(b *strings.Builder, rows []Input) {
	b.WriteString("## Ingress Passthroughs\n\n")
	for _, row := range rows {
		fmt.Fprintf(b, "- `%s`: %s Evidence: %s.\n", row.Name, row.Summary, checks(row.SourceChecks))
	}
	b.WriteByte('\n')
}

func renderAbsentScans(b *strings.Builder, rows []AbsentScan) {
	b.WriteString("## Forbidden Inference Scans\n\n")
	b.WriteString("| Name | Scope | Forbidden tokens | Reason |\n| --- | --- | --- | --- |\n")
	for _, row := range rows {
		fmt.Fprintf(b, "| %s | %s | %s | %s |\n",
			row.Name, checks(row.Scope), checks(row.Tokens), row.Reason)
	}
	b.WriteByte('\n')
}

func renderAssertions(b *strings.Builder, assertions []string) {
	b.WriteString("## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- %s\n", assertion)
	}
}
