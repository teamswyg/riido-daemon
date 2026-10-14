package main

import (
	"fmt"
	"strings"
)

func executionIdentityDoc(m model) string {
	return renderDoc("Execution Identity", "Overview", "overview.md", func(b *strings.Builder) {
		b.WriteString("ExecutionIdentity should be shared vocabulary in riido-contracts.\n\n")
		b.WriteString("| Field | Meaning |\n| --- | --- |\n")
		for _, row := range m.Execution.IdentityFields {
			fmt.Fprintf(b, "| `%s` | %s |\n", row.Field, row.Meaning)
		}
		b.WriteString("\nRules:\n\n")
		writeBullets(b, m.Execution.IdentityRules)
	})
}

func workspacePlanDoc(m model) string {
	return renderDoc("Workspace Plan", "Overview", "overview.md", func(b *strings.Builder) {
		b.WriteString("F3 is fixed by passing structured assignment snapshot fields, not prompt text.\n\n")
		b.WriteString("| Field | Phase | Rule |\n| --- | --- | --- |\n")
		for _, row := range m.Execution.WorkspaceFields {
			fmt.Fprintf(b, "| `%s` | %s | %s |\n", row.Field, row.Phase, row.Rule)
		}
		b.WriteString("\nP0 clones public repositories only; private repositories fail closed until token-ref broker ownership is defined.\n")
	})
}

func launchEnvelopeDoc(m model) string {
	return renderDoc("Runtime Launch Envelope", "Overview", "overview.md", func(b *strings.Builder) {
		b.WriteString("Provider detection and process spawn are separate phases.\n\n")
		b.WriteString("| Field | Owner | Rule |\n| --- | --- | --- |\n")
		for _, row := range m.Execution.LaunchFields {
			fmt.Fprintf(b, "| `%s` | %s | %s |\n", row.Field, row.Owner, row.Rule)
		}
	})
}
