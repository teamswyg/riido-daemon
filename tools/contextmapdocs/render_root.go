package main

import (
	"fmt"
	"strings"
)

func renderRoot(m manifest) string {
	return renderDoc(m.Title, "", func(b *strings.Builder) {
		fmt.Fprintf(b, "> Riido task: %s `%s`\n\n", m.RiidoTask, m.TaskTitle)
		fmt.Fprintf(b, "%s\n\n", m.Intro)
		writeBullets(b, m.OwnershipSummary)
		b.WriteString("\nFocused sections:\n\n")
		for _, link := range m.FocusedSections {
			fmt.Fprintf(b, "- [%s](%s)\n", link.Title, link.Path)
		}
		b.WriteString("\nBoundary evidence:\n\n")
		for _, link := range m.BoundaryEvidence {
			fmt.Fprintf(b, "- [%s](%s)\n", link.Title, link.Path)
		}
		fmt.Fprintf(b, "\nRepository boundary rule: %s\n\n", m.RepositoryRule)
		fmt.Fprintf(b, "Work-unit boundary rule: %s The executable PR gate is `%s`, backed by `%s`.\n",
			m.WorkUnitRule, m.WorkUnitGate, m.WorkUnitScript)
	})
}
