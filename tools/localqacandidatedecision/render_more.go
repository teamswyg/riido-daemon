package main

import (
	"fmt"
	"strings"
)

func renderDecisionTable(b *strings.Builder, decisions []decisionRecord) {
	fmt.Fprintln(b, "## Decisions")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Candidate | Disposition | Priority | Owner | Review By | Next Artifact |")
	fmt.Fprintln(b, "| --- | --- | --- | --- | --- | --- |")
	for _, decision := range decisions {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | `%s` | `%s` | `%s` |\n",
			decision.CandidateID, decision.Disposition, decision.Priority,
			decision.Owner, decision.ReviewBy, decision.NextArtifact)
	}
	fmt.Fprintln(b)
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	fmt.Fprintln(b, "## Loop")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "- Observation: %s\n", loop.Observation)
	fmt.Fprintf(b, "- Hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- Execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- Evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- Retrospective: %s\n", loop.Retrospective)
}
