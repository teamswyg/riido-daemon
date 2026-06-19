package main

import (
	"fmt"
	"strings"
)

func verificationEvidenceDoc(m model) string {
	return renderDoc("Verification Evidence", "Assignment Lifecycle FSM", "assignment-lifecycle-fsm.md", func(b *strings.Builder) {
		b.WriteString("The executable evidence manifest is verified by tools/agentexecutionevidence.\n\n")
		b.WriteString("| Risk | Evidence | Proves |\n| --- | --- | --- |\n")
		for _, item := range m.Items {
			fmt.Fprintf(b, "| `%s` | `%s` | %s |\n", item.Risk, evidenceRef(item), item.Proves)
		}
	})
}

func currentStatusDoc(m model) string {
	return renderDoc("Current Daemon Slice Status", "Assignment Lifecycle FSM", "assignment-lifecycle-fsm.md", func(b *strings.Builder) {
		fmt.Fprintf(b, "Verified evidence items: `%d`.\n\n", len(m.Items))
		b.WriteString("| Risk | Status | Evidence owner |\n| --- | --- | --- |\n")
		for _, item := range m.Items {
			fmt.Fprintf(b, "| `%s` | `%s` | `%s` |\n", item.Risk, item.Status, ownerRef(item))
		}
		b.WriteString("\nRemaining boundaries:\n\n")
		b.WriteString("| ID | Owner | Current handling | Required next artifact |\n| --- | --- | --- | --- |\n")
		for _, item := range m.Boundaries {
			fmt.Fprintf(b, "| `%s` | %s | %s | %s |\n", item.ID, item.Owner, item.CurrentHandling, item.RequiredNextArtifact)
		}
	})
}

func evidenceRef(item evidenceItem) string {
	if item.Test == "" {
		return ownerRef(item)
	}
	return ownerRef(item) + ":" + item.Test
}

func ownerRef(item evidenceItem) string {
	if item.Repo != "" {
		return item.Repo
	}
	return item.Package
}
