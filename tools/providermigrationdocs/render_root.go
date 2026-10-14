package main

import (
	"fmt"
	"strings"
)

func renderRoot(m manifest) string {
	return renderDoc(m.Title, "", func(b *strings.Builder) {
		fmt.Fprintf(b, "Executable SSOT: [`public-migration-status.riido.json`](public-migration-status.riido.json).\n\n")
		fmt.Fprintf(b, "Evidence artifact: `%s`.\n\n", m.EvidenceArtifact)
		writeParagraphs(b, m.Summary)
		b.WriteString("Focused status files:\n\n")
		for _, page := range m.Pages {
			fmt.Fprintf(b, "- [%s](public-migration-status/%s.md)\n", page.Title, page.ID)
		}
		b.WriteString("\nAssertions:\n\n")
		writeBullets(b, m.Assertions)
	})
}
