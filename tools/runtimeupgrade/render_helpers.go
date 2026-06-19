package main

import (
	"bytes"
	"fmt"
	"strings"
)

func ruleSummary(rule Rule) string {
	if rule.Summary != "" {
		return rule.Summary
	}
	return rule.Owner
}

func ruleEvidence(rule Rule) string {
	if rule.Status == "reserved" {
		return "requires: " + rule.RequiredEvidence
	}
	return "`" + strings.Join(rule.SourceChecks, "`, `") + "`"
}

func ruleDecisions(rule Rule) string {
	if len(rule.DecisionRefs) == 0 {
		return ""
	}
	return "`" + strings.Join(rule.DecisionRefs, "`, `") + "`"
}

func writeAssertions(b *bytes.Buffer, assertions []string) {
	b.WriteString("## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- `%s`\n", assertion)
	}
}
