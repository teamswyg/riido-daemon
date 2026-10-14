package main

import (
	"bytes"
	"fmt"
)

func writeAbsentSurfaces(b *bytes.Buffer, surfaces []AbsentSurface) {
	b.WriteString("\n## Absent Surfaces\n\n| Surface | Reason |\n| --- | --- |\n")
	for _, surface := range surfaces {
		fmt.Fprintf(b, "| `%s` | %s |\n", surface.Name, surface.Reason)
	}
}

func writeAssertions(b *bytes.Buffer, assertions []string) {
	b.WriteString("\n## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- `%s`\n", assertion)
	}
}
