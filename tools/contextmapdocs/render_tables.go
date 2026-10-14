package main

import (
	"fmt"
	"strings"
)

func renderBoundedContexts(m manifest) string {
	return renderDoc("Bounded Contexts", "../context-map.md", func(b *strings.Builder) {
		b.WriteString("| ID | Context | Public daemon owner |\n| --- | --- | --- |\n")
		for _, row := range m.Contexts {
			fmt.Fprintf(b, "| %s | %s | %s |\n", row.ID, row.Context, row.Owner)
		}
	})
}

func renderACL(m manifest) string {
	return renderDoc("ACL Locations", "../context-map.md", func(b *strings.Builder) {
		b.WriteString("| ACL | Input | Output |\n| --- | --- | --- |\n")
		for _, row := range m.ACL.Rows {
			fmt.Fprintf(b, "| %s | %s | %s |\n", row.ACL, row.Input, row.Output)
		}
	})
}

func renderDependency(m manifest) string {
	return renderDoc("Dependency Direction", "../context-map.md", func(b *strings.Builder) {
		b.WriteString("```text\n")
		for _, line := range m.Dependency.Diagram {
			fmt.Fprintln(b, line)
		}
		b.WriteString("```\n\n")
		writeBullets(b, m.Dependency.Notes)
		fmt.Fprintf(b, "\nRuntime/domain packages must not import %s.\n\n", inlineCodeList(m.Dependency.ForbiddenImports))
		fmt.Fprintln(b, m.Dependency.RetiredPrivateRepoRule)
	})
}
