package main

import (
	"bytes"
	"fmt"
)

func writeSourceChecks(b *bytes.Buffer, checks []SourceCheck) {
	b.WriteString("\n## Source Evidence\n\n| Check | Source |\n| --- | --- |\n")
	for _, check := range checks {
		fmt.Fprintf(b, "| %s | `%s` |\n", check.Name, check.File)
	}
}

func writeAssertions(b *bytes.Buffer, assertions []string) {
	b.WriteString("\n## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- `%s`\n", assertion)
	}
}
