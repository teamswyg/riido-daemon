package main

import (
	"fmt"
	"strings"
)

func renderRoot(m manifest) string {
	return renderDoc(m.Title, "", func(b *strings.Builder) {
		b.WriteString("Executable SSOT: [`release-artifacts.riido.json`](release-artifacts.riido.json).\n\n")
		fmt.Fprintf(b, "Evidence artifact: `%s`.\n\n%s\n\n", m.EvidenceArtifact, m.Purpose)
		b.WriteString("## Release Surfaces\n\n")
		for _, doc := range m.DetailDocs {
			fmt.Fprintf(b, "- [%s](%s)\n", doc.Title, pathFromRoot(doc.Path))
		}
		b.WriteString("\n## Executable Sources\n\n")
		writeBullets(b, []string{
			m.ReleaseWorkflow,
			m.CDNPublishWorkflow,
			m.BuildScript,
			m.PublishScript,
			m.CDNPublishScript,
			m.InstallScript,
		})
		b.WriteString("## Assertions\n\n")
		writeBullets(b, m.Assertions)
	})
}

func pathFromRoot(path string) string {
	return strings.TrimPrefix(path, "docs/30-architecture/")
}
