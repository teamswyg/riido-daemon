package main

import (
	"fmt"
	"strings"
)

func renderManifestLoopInventory(b *strings.Builder, e evidence) {
	b.WriteString("## Manifest Loop Inventory\n\n")
	fmt.Fprintf(b, "- Complete manifest loops: `%d`\n", e.ManifestLoopCount)
	fmt.Fprintf(b, "- Direct manifest loops: `%d`\n", e.ManifestDirectLoopCount)
	fmt.Fprintf(b, "- Delegated manifest loops: `%d`\n", e.ManifestDelegatedLoopCount)
	fmt.Fprintf(b, "- Missing manifest loops: `%d`\n", e.ManifestMissingLoopCount)
	fmt.Fprintf(b, "- Missing loop budget: `%d`\n\n", e.ManifestLoopBudget.MaxMissing)
	b.WriteString("| Group | Missing loops | Budget | Sample paths |\n| --- | ---: | ---: | --- |\n")
	if len(e.ManifestMissingLoopGroups) == 0 {
		b.WriteString("| None | 0 | 0 | - |\n\n")
		return
	}
	for _, group := range e.ManifestMissingLoopGroups {
		fmt.Fprintf(b, "| `%s` | %d | %d | %s |\n", group.Group, group.Count, loopGroupBudget(e, group.Group), loopSampleText(e, group.Group))
	}
	b.WriteString("\n")
}

func loopGroupBudget(e evidence, group string) int {
	return e.ManifestLoopBudget.MaxMissingByGroup[group]
}

func loopSampleText(e evidence, group string) string {
	for _, sample := range e.ManifestMissingLoopSamples {
		if sample.Group == group {
			return samplePathText(sample.Paths)
		}
	}
	return "None"
}
