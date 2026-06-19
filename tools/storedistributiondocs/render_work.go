package main

import (
	"fmt"
	"strings"
)

func workTableDoc(title string, rows []workRow) string {
	return renderDoc(title, "Daemon Changes", func(b *strings.Builder) {
		b.WriteString("| Work | Owner context | Output |\n| --- | --- | --- |\n")
		for _, row := range rows {
			fmt.Fprintf(b, "| %s | %s | %s |\n", row.Work, row.OwnerContext, row.Output)
		}
	})
}

func reviewNotesDoc(m manifest, c contract) string {
	return renderDoc("Review Notes Contract", "Daemon Changes", func(b *strings.Builder) {
		b.WriteString("Store submission notes must state:\n\n")
		for i, note := range m.ReviewNotes {
			fmt.Fprintf(b, "%d. %s\n", i+1, note)
		}
		b.WriteString("\nProvider CLIs covered by this contract:\n\n")
		writeBullets(b, c.ExternalProviderCLINames)
		b.WriteString("\nRequired store-review surfaces:\n\n")
		writeBullets(b, storeReviewSurfaces(c))
	})
}

func storeReviewSurfaces(c contract) []string {
	seen := map[string]bool{}
	for _, item := range c.Channels {
		for _, surface := range item.RequiredSurfaces {
			if strings.Contains(surface, "review") || strings.Contains(surface, "privacy") {
				seen[surface] = true
			}
		}
	}
	return sortedKeys(seen)
}
