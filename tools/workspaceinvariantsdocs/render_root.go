package main

import (
	"fmt"
	"strings"
)

func renderRoot(m manifest) string {
	return renderDoc(m.Title, "", func(b *strings.Builder) {
		b.WriteString("[Back to workspace](../workspace.md)\n\n")
		fmt.Fprintf(b, "Executable SSOT: [`invariants.riido.json`](invariants.riido.json).\n\n")
		fmt.Fprintf(b, "Evidence artifact: `%s`.\n\n", m.EvidenceArtifact)
		for _, summary := range m.Summary {
			fmt.Fprintf(b, "%s\n\n", summary)
		}
		b.WriteString("## Generated Detail Surfaces\n\n")
		for _, link := range m.DetailPages {
			fmt.Fprintf(b, "- [%s](%s)\n", link.Title, link.Path)
		}
		b.WriteString("\n## Related Evidence Surfaces\n\n")
		for _, link := range m.RelatedPages {
			fmt.Fprintf(b, "- [%s](%s)\n", link.Title, link.Path)
		}
		b.WriteString("\n## Executable Assertions\n\n")
		writeBullets(b, m.Assertions)
	})
}
