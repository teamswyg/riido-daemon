package main

import (
	"fmt"
	"strings"
)

func renderRoot(m manifest) string {
	return renderDoc(m.Title, func(b *strings.Builder) {
		b.WriteString("Executable SSOT: [`security.riido.json`](security.riido.json).\n\n")
		fmt.Fprintf(b, "Evidence artifact: `%s`.\n\n", m.EvidenceArtifact)
		for _, summary := range m.Summary {
			fmt.Fprintf(b, "%s\n\n", summary)
		}
		b.WriteString("## Compatibility Markers\n\n")
		writeHarness(b, m)
		writeBullets(b, m.CompatibilityMarkers)
		b.WriteString("\n## Parts\n\n")
		writeLinks(b, m.Parts)
		b.WriteString("\n## Related Evidence Surfaces\n\n")
		writeLinks(b, m.RelatedPages)
		b.WriteString("\n## Executable Assertions\n\n")
		writeBullets(b, m.Assertions)
	})
}

func writeHarness(b *strings.Builder, m manifest) {
	if m.FullAccessHarnessTitle == "" {
		return
	}
	fmt.Fprintf(b, "### %s\n\n", m.FullAccessHarnessTitle)
	for _, paragraph := range m.FullAccessHarness {
		fmt.Fprintf(b, "%s\n\n", paragraph)
	}
	if m.FullAccessHarnessCode != "" {
		fmt.Fprintf(b, "```text\n%s\n```\n\n", m.FullAccessHarnessCode)
	}
}

func writeLinks(b *strings.Builder, links []link) {
	for _, link := range links {
		fmt.Fprintf(b, "- [%s](%s)\n", link.Title, link.Path)
	}
}
