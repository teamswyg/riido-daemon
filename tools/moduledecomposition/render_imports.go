package main

import (
	"fmt"
	"strings"
)

func renderImportRules(m manifest) string {
	return renderDoc("Import Rules", "../module-decomposition.md", func(b *strings.Builder) {
		b.WriteString("| Package group | May import | Must not import |\n| --- | --- | --- |\n")
		for _, rule := range m.ImportRules {
			fmt.Fprintf(b, "| `%s` | %s | %s |\n", rule.Group, rule.MayImport, rule.MustNotImport)
		}
		b.WriteString("\nProduction imports are checked with `go list`; tests may use package-local fakes.\n\n")
		b.WriteString("Checked forbidden prefixes:\n\n")
		for _, rule := range m.ImportRules {
			fmt.Fprintf(b, "- `%s`: %s\n", rule.Group, codeList(rule.ForbiddenPrefixes))
		}
	})
}
