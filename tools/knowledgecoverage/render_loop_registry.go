package main

import (
	"fmt"
	"strings"
)

func renderLoopRegistry(b *strings.Builder, entries []loopRegistryEntry) {
	b.WriteString("\n## Loop Registry\n\n")
	if len(entries) == 0 {
		b.WriteString("_None._\n")
		return
	}
	b.WriteString("| Loop | Expires | Observes | Verifies | Evidence | Fails when |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	for _, entry := range entries {
		fmt.Fprintf(
			b,
			"| `%s` | `%s` | %s | %s | %s | %s |\n",
			entry.ID,
			entry.ExpiresAfter,
			joinList(entry.Observes),
			joinList(entry.Verifies),
			joinList(entry.Evidence),
			joinList(entry.FailsWhen),
		)
	}
}

func joinList(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, "<br>")
}
