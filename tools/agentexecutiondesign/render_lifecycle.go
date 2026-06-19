package main

import (
	"fmt"
	"strings"
)

func streamEnvelopeDoc(m model) string {
	return renderDoc("Stream Envelope", "Assignment Lifecycle FSM", "assignment-lifecycle-fsm.md", func(b *strings.Builder) {
		b.WriteString("| Event kind | Store | Client meaning |\n| --- | --- | --- |\n")
		for _, row := range m.Lifecycle.StreamEvents {
			fmt.Fprintf(b, "| `%s` | %s | %s |\n", row.Kind, row.Store, row.ClientMeaning)
		}
		fmt.Fprintf(b, "\nRule: %s\n", m.Lifecycle.StreamRule)
	})
}

func retryPolicyDoc(m model) string {
	return renderDoc("Retry And Recovery Policy", "Assignment Lifecycle FSM", "assignment-lifecycle-fsm.md", func(b *strings.Builder) {
		b.WriteString("| Error class | Retry | Rule |\n| --- | --- | --- |\n")
		for _, row := range m.Lifecycle.RetryPolicies {
			fmt.Fprintf(b, "| `%s` | %s | %s |\n", row.Class, row.Retry, row.Rule)
		}
	})
}

func implementationSlicesDoc(m model) string {
	return renderDoc("Implementation Slices", "Assignment Lifecycle FSM", "assignment-lifecycle-fsm.md", func(b *strings.Builder) {
		for _, slice := range m.Lifecycle.ImplementationSlices {
			fmt.Fprintf(b, "## %s\n\n", slice.Title)
			writeBullets(b, slice.Items)
			b.WriteString("\n")
		}
	})
}
