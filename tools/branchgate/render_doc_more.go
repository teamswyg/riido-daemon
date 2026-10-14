package main

import (
	"bytes"
	"fmt"
)

func writeExamples(b *bytes.Buffer, examples []BranchExample) {
	b.WriteString("\n## Examples\n\n| Branch | Accepted | Reason |\n| --- | --- | --- |\n")
	for _, ex := range examples {
		fmt.Fprintf(b, "| `%s` | `%t` | %s |\n", ex.Branch, ex.Accepted, ex.Reason)
	}
}

func writeAssertions(b *bytes.Buffer, assertions []string) {
	b.WriteString("\n## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- `%s`\n", assertion)
	}
}
