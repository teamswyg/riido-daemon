package main

import (
	"fmt"
	"strings"
)

func renderSummary(b *strings.Builder, m manifest) {
	fmt.Fprintf(b, "- loop count: `%d`\n", len(m.Loops))
	fmt.Fprintf(b, "- registered loop files: `%d`\n", len(m.LoopFiles))
	fmt.Fprintf(b, "- evidence items: `%d`\n", evidenceItemCount(m.Loops))
	fmt.Fprintf(b, "- open gaps: `%d`\n", len(m.OpenGaps))
	b.WriteString("- phase coverage: ")
	for i, row := range phaseCoverageRows(m) {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "`%s=%d/%d`", row.Phase, row.Count, len(m.Loops))
	}
	b.WriteString("\n\n")
}
