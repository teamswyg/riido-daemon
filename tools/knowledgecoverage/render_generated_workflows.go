package main

import (
	"fmt"
	"strings"
)

func renderGeneratedWorkflowCoverage(
	b *strings.Builder,
	coverage generatedOriginWorkflowCoverage,
) {
	b.WriteString("## Generated Origin Workflow Coverage\n\n")
	fmt.Fprintf(b, "- Covered generated origins: `%d`\n", coverage.CoveredCount)
	fmt.Fprintf(b, "- Missing workflow origins: `%d`\n\n", coverage.MissingCount)
	if len(coverage.Missing) == 0 {
		b.WriteString("_None missing._\n\n")
		return
	}
	b.WriteString("| Generator | Tool | Count |\n| --- | --- | ---: |\n")
	for _, missing := range coverage.Missing {
		fmt.Fprintf(b, "| `%s` | `%s` | %d |\n", missing.Generator, missing.Tool, missing.Count)
	}
	b.WriteString("\n")
}
