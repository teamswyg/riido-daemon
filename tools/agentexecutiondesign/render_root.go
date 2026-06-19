package main

import (
	"fmt"
	"strings"
)

func rootDoc(m model) string {
	return renderDoc("Agent Execution Unresolved Design", "", "", func(b *strings.Builder) {
		fmt.Fprintf(b, "Executable SSOT: [`agent-execution-unresolved-design.riido.json`](agent-execution-unresolved-design.riido.json).\n\n")
		fmt.Fprintf(b, "Evidence artifact: `%s`.\n\n", m.Manifest.EvidenceArtifact)
		b.WriteString("## Parts\n\n")
		for _, link := range m.Overview.FocusedFiles {
			fmt.Fprintf(b, "- [%s](agent-execution-unresolved-design/%s)\n", link.Title, link.Path)
		}
	})
}

func overviewDoc(m model) string {
	return renderDoc("Agent Execution Unresolved Design: Overview", "agent-execution-unresolved-design.md", "../agent-execution-unresolved-design.md", func(b *strings.Builder) {
		fmt.Fprintf(b, "Riido task: %s.\n\n", m.Manifest.RiidoTask)
		b.WriteString("This SSOT captures daemon-side unresolved agent execution risks as executable architecture inputs.\n\n")
		b.WriteString("The shared shape is:\n\n")
		writeNumbered(b, m.Overview.SharedShape)
		b.WriteString("\nFocused model files:\n\n")
		for _, link := range m.Overview.FocusedFiles {
			fmt.Fprintf(b, "- [%s](%s)\n", link.Title, link.Path)
		}
	})
}
