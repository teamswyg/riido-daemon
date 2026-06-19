package main

import (
	"fmt"
	"strings"
)

func renderCore(m manifest) string {
	return renderDoc(m.Core.Title, func(b *strings.Builder) {
		b.WriteString("[Back to invariants](../invariants.md)\n\n")
		fmt.Fprintf(b, "Executable SSOT: [`runtime-scheduling.riido.json`](../../runtime-scheduling.riido.json).\n\n")
		fmt.Fprintf(b, "Core fragment: [`core.riido.json`](core.riido.json).\n\n")
		fmt.Fprintf(b, "%s\n\n", m.Core.Context)
		b.WriteString("## Responsibilities\n\n")
		writeBullets(b, m.Core.Responsibilities)
		b.WriteString("\n## Non-Responsibilities\n\n")
		writeBullets(b, m.Core.NonResponsibilities)
		b.WriteString("\n## Core Invariants\n\n")
		writeInvariantTable(b, m.Core.Invariants)
	})
}

func writeInvariantTable(b *strings.Builder, invariants []invariant) {
	b.WriteString("| Invariant | Rule | Evidence |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, inv := range invariants {
		fmt.Fprintf(b, "| %s | %s | `%s` |\n", inv.Name, inv.Summary, strings.Join(inv.SourceChecks, "`, `"))
	}
}
