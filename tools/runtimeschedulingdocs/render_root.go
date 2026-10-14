package main

import (
	"fmt"
	"strings"
)

func renderRoot(m manifest) string {
	return renderDoc(m.Title, func(b *strings.Builder) {
		fmt.Fprintf(b, "Executable SSOT: [`runtime-scheduling.riido.json`](runtime-scheduling.riido.json).\n\n")
		fmt.Fprintf(b, "Evidence artifact: `%s`.\n\n", m.EvidenceArtifact)
		for _, summary := range m.Summary {
			fmt.Fprintf(b, "%s\n\n", summary)
		}
		b.WriteString("## Parts\n\n")
		writeLinks(b, m.Parts)
		b.WriteString("\n## Evidence Surfaces\n\n")
		writeLinks(b, m.EvidencePages)
		b.WriteString("\n## Executable Assertions\n\n")
		writeBullets(b, m.Assertions)
	})
}

func writeLinks(b *strings.Builder, links []link) {
	for _, link := range links {
		fmt.Fprintf(b, "- [%s](%s)\n", link.Title, link.Path)
	}
}
