package main

import (
	"bytes"
	"fmt"
)

func writeAssertions(b *bytes.Buffer, assertions []string) {
	b.WriteString("\n## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- `%s`\n", assertion)
	}
}
