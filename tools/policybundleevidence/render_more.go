package main

import (
	"fmt"
	"strings"
)

func renderAbsent(b *strings.Builder, rows []AbsentSurface) {
	b.WriteString("## Forbidden Surfaces\n\n| Surface | Reason |\n| --- | --- |\n")
	for _, row := range rows {
		fmt.Fprintf(b, "| `%s` | %s |\n", row.Name, row.Reason)
	}
	b.WriteByte('\n')
}

func renderAssertions(b *strings.Builder, rows []string) {
	b.WriteString("## Assertions\n\n")
	for _, row := range rows {
		fmt.Fprintf(b, "- `%s`\n", row)
	}
}
