package main

import (
	"fmt"
	"strings"
)

func repoOwnershipDoc(m model) string {
	return renderDoc("Repo Ownership", "Assignment Lifecycle FSM", "assignment-lifecycle-fsm.md", func(b *strings.Builder) {
		b.WriteString("| Repo | SSOT responsibility |\n| --- | --- |\n")
		for _, row := range m.Governance.RepoOwnership {
			fmt.Fprintf(b, "| `%s` | %s |\n", row.Repo, row.Responsibility)
		}
		b.WriteString("\nIf vocabulary is needed in two or more repos, prefer promotion to riido-contracts.\n")
	})
}

func ragGuardrailsDoc(m model) string {
	return renderDoc("Agent Execution Unresolved Design: RAG Guardrails", "agent-execution-unresolved-design.md", "../agent-execution-unresolved-design.md", func(b *strings.Builder) {
		b.WriteString("Public-safe RAG may contain:\n\n")
		writeBullets(b, m.Governance.RAGAllowed)
		b.WriteString("\nRAG index must not contain:\n\n")
		writeBullets(b, m.Governance.RAGForbidden)
		b.WriteString("\n## Open Decisions\n\n")
		b.WriteString("| ID | Decision needed | Default until decided |\n| --- | --- | --- |\n")
		for _, row := range m.Governance.OpenDecisions {
			fmt.Fprintf(b, "| %s | %s | %s |\n", row.ID, row.Decision, row.Default)
		}
		b.WriteString("\n## Non-goals\n\n")
		writeBullets(b, m.Governance.NonGoals)
	})
}
