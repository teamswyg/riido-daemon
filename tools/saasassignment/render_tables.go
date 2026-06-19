package main

import (
	"bytes"
	"fmt"
)

func writeFactTable(b *bytes.Buffer, facts []Fact) {
	b.WriteString("## Implemented Facts\n\n")
	b.WriteString("| Fact | Status | Summary | Evidence |\n| --- | --- | --- | --- |\n")
	for _, fact := range facts {
		fmt.Fprintf(b, "| %s | `%s` | %s | %s |\n", fact.Name, fact.Status, fact.Summary, checksList(fact.SourceChecks))
	}
	b.WriteString("\n")
}

func writeBoundaryTable(b *bytes.Buffer, boundaries []Boundary) {
	b.WriteString("## Boundaries\n\n")
	b.WriteString("| Boundary | Owner | Summary |\n| --- | --- | --- |\n")
	for _, boundary := range boundaries {
		fmt.Fprintf(b, "| %s | %s | %s |\n", boundary.Name, boundary.Owner, boundary.Summary)
	}
	b.WriteString("\n")
}

func writeAbsentTable(b *bytes.Buffer, surfaces []AbsentSurface) {
	b.WriteString("## Forbidden Surfaces\n\n")
	b.WriteString("| Surface | Reason |\n| --- | --- |\n")
	for _, surface := range surfaces {
		fmt.Fprintf(b, "| `%s` | %s |\n", surface.Name, surface.Reason)
	}
	b.WriteString("\n")
}
