package main

import (
	"bytes"
	"fmt"
)

func writeLevels(b *bytes.Buffer, levels []Level, timeouts []Timeout) {
	b.WriteString("\n## Shutdown Authority\n\n")
	b.WriteString("Shutdown authority is owned by `pkg/lifecycle`.\n\n")
	b.WriteString("| Level | Const | Order |\n| --- | --- | --- |\n")
	for _, row := range levels {
		fmt.Fprintf(b, "| `%s` | `%s` | `%d` |\n", row.Name, row.Const, row.Order)
	}
	b.WriteString("\n| Timeout const | Duration |\n| --- | --- |\n")
	for _, row := range timeouts {
		fmt.Fprintf(b, "| `%s` | `%s` |\n", row.Const, row.Duration)
	}
}

func writeAssertions(b *bytes.Buffer, assertions []string) {
	b.WriteString("\n## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- `%s`\n", assertion)
	}
}
