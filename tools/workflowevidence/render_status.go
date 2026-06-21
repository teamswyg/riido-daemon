package main

import (
	"fmt"
	"strings"
)

func renderStatusCounts(b *strings.Builder, records []workflowRecord) {
	b.WriteString("- workflow statuses: ")
	for i, row := range workflowStatusCounts(records) {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "`%s=%d`", row.Status, row.Count)
	}
	b.WriteString("\n")
}
