package main

import (
	"fmt"
	"strings"
)

func renderRoot(m manifest) string {
	return renderDoc(m.Title, "", func(b *strings.Builder) {
		fmt.Fprintf(b, "> Riido task: %s\n\n", m.RiidoTask)
		b.WriteString("Executable SSOT: [`integration-matrix.riido.json`](integration-matrix.riido.json).\n\n")
		fmt.Fprintf(b, "Evidence artifact: `%s`.\n\n%s\n\n", m.EvidenceArtifact, m.Purpose)
		fmt.Fprintf(b, "Provider evidence is loaded from [`%s`](provider-validation-matrix.riido.json).\n\n", m.ProviderValidationManifest)
		fmt.Fprintf(b, "The security decision itself is owned by [`security.md`](%s) §4.3.\n\n", m.SecurityDecisionLink)
		b.WriteString("## Matrix Surfaces\n\n")
		for _, doc := range m.DetailDocs {
			fmt.Fprintf(b, "- [%s](%s)\n", doc.Title, rootLink(doc.Path))
		}
	})
}

func rootLink(path string) string {
	return strings.TrimPrefix(path, "docs/30-architecture/")
}
