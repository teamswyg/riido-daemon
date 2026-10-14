package main

import (
	"fmt"
	"strings"
)

func rootDoc(m manifest) string {
	return renderDoc(m.Title, "", func(b *strings.Builder) {
		fmt.Fprintf(b, "Executable SSOT: [`store-distribution.riido.json`](store-distribution.riido.json).\n\n")
		fmt.Fprintf(b, "Evidence artifact: `%s`.\n\n", m.EvidenceArtifact)
		b.WriteString("## Compatibility Markers\n\n")
		writeBullets(b, m.CompatibilityMarkers)
		b.WriteString("\n## Parts\n\n")
		b.WriteString("- [Architecture](store-distribution/architecture.md)\n")
		b.WriteString("- [Daemon Changes](store-distribution/daemon-changes.md)\n")
	})
}

func architectureIndexDoc(m manifest) string {
	return renderDoc("Store Distribution Architecture SSOT: Architecture", "store-distribution.md", func(b *strings.Builder) {
		fmt.Fprintf(b, "Store contract: [`%s`](../../../%s).\n\n", m.StoreContract, m.StoreContract)
		b.WriteString("Focused sections:\n\n")
		for _, item := range architectureLinks {
			fmt.Fprintf(b, "- [%s](architecture/%s)\n", item.Title, item.File)
		}
	})
}

func daemonIndexDoc() string {
	return renderDoc("Store Distribution Architecture SSOT: Daemon Changes", "store-distribution.md", func(b *strings.Builder) {
		b.WriteString("## Detail Surfaces\n\n")
		for _, item := range daemonLinks {
			fmt.Fprintf(b, "- [%s](daemon-changes/%s)\n", item.Title, item.File)
		}
	})
}
