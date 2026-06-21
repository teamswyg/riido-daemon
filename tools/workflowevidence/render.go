package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, result auditResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	fmt.Fprintf(&b, "- workflow count: `%d`\n", len(result.Records))
	fmt.Fprintf(&b, "- evidence covered: `%d`\n", result.Covered)
	fmt.Fprintf(&b, "- evidence tools covered: `%d/%d`\n",
		result.EvidenceToolCovered, result.EvidenceTools)
	fmt.Fprintf(&b, "- evidence tools bound to artifacts: `%d/%d`\n",
		result.EvidenceToolBound, result.EvidenceTools)
	fmt.Fprintf(&b, "- accepted gaps: `%d`\n", result.Accepted)
	fmt.Fprintf(&b, "- non-strict artifact uploads: `%d`\n", len(result.NonStrict))
	fmt.Fprintf(&b, "- missing evidence uploads: `%d`\n", len(result.MissingEvidence))
	fmt.Fprintf(&b, "- unregistered gaps: `%d`\n\n", len(result.Unregistered))
	renderLoop(&b, m.Loop)
	renderWorkflowTable(&b, result.Records)
	renderAssertions(&b, m.Assertions)
	return b.String()
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	b.WriteString("## Evidence Loop\n\n")
	b.WriteString("| Step | Evidence |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Observe | %s |\n", loop.Observation)
	fmt.Fprintf(b, "| Hypothesis | %s |\n", loop.Hypothesis)
	fmt.Fprintf(b, "| Execute | %s |\n", loop.Execute)
	fmt.Fprintf(b, "| Evaluate | %s |\n", loop.Evaluate)
	fmt.Fprintf(b, "| Retrospective | %s |\n\n", loop.Retrospective)
}
