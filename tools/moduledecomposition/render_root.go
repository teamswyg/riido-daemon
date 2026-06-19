package main

import (
	"fmt"
	"strings"
)

func renderRoot(m manifest) string {
	return renderDoc(m.Title, "", func(b *strings.Builder) {
		fmt.Fprintf(b, "> Riido task: %s\n\n", m.RiidoTask)
		b.WriteString("Executable SSOT: [`module-decomposition.riido.json`](module-decomposition.riido.json).\n\n")
		fmt.Fprintf(b, "Evidence artifact: `%s`.\n\n%s\n\n", m.EvidenceArtifact, m.Purpose)
		b.WriteString("## Decisions\n\n")
		for i, decision := range m.Decisions {
			fmt.Fprintf(b, "%d. %s\n", i+1, decision)
		}
		b.WriteString("\n## SSOT Fragments\n\n")
		writeFragmentLink(b, "Package roles", m.PackageRolesFile)
		writeFragmentLink(b, "Import rules", m.ImportRulesFile)
		writeFragmentLink(b, "Ports", m.PortsFile)
		b.WriteString("\n## Detail Surfaces\n\n")
		for _, doc := range m.DetailDocs {
			fmt.Fprintf(b, "- [%s](%s)\n", doc.Title, rootLink(doc.Path))
		}
		b.WriteString("\nThe user/operator command boundary is owned by [`cli-surface.md`](cli-surface.md).\n")
	})
}

func rootLink(path string) string {
	return strings.TrimPrefix(path, "docs/30-architecture/")
}

func writeFragmentLink(b *strings.Builder, label, path string) {
	if path == "" {
		return
	}
	fmt.Fprintf(b, "- %s: [`%s`](%s)\n", label, path, path)
}
