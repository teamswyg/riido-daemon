package main

import (
	"fmt"
	"strings"
)

func renderRoot(m manifest) string {
	return renderDoc(m.Title, "", func(b *strings.Builder) {
		fmt.Fprintf(b, "Executable SSOT: [`locking.riido.json`](locking.riido.json).\n\n")
		fmt.Fprintf(b, "Evidence artifact: `%s`.\n\n", m.EvidenceArtifact)
		for _, summary := range m.Summary {
			fmt.Fprintf(b, "%s\n\n", summary)
		}
		b.WriteString("## Responsibilities\n\n")
		writeBullets(b, m.Responsibilities)
		b.WriteString("\n## Non-Responsibilities\n\n")
		writeBullets(b, m.NonResponsibilities)
		b.WriteString("\n## Core Invariants\n\n")
		for i, item := range m.Invariants {
			fmt.Fprintf(b, "%d. %s\n", i+1, item)
		}
		b.WriteString("\n## Detail Surfaces\n\n")
		for _, link := range m.DetailPages {
			fmt.Fprintf(b, "- [%s](%s)\n", link.Title, link.Path)
		}
		b.WriteString("\n## Assertions\n\n")
		writeBullets(b, m.Assertions)
	})
}
