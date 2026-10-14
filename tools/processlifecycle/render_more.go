package main

import (
	"bytes"
	"fmt"
	"strings"
)

func writeInterfaces(b *bytes.Buffer, interfaces []InterfaceSpec) {
	b.WriteString("\n## Interface Surface\n\n| Interface | Methods |\n| --- | --- |\n")
	for _, row := range interfaces {
		fmt.Fprintf(b, "| `%s` | `%s` |\n", row.Name, strings.Join(row.Methods, "`, `"))
	}
}

func writeAssertions(b *bytes.Buffer, assertions []string) {
	b.WriteString("\n## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- `%s`\n", assertion)
	}
}
