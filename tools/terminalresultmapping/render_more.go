package main

import (
	"bytes"
	"fmt"
)

func writeLifecycle(b *bytes.Buffer) {
	b.WriteString("\n## Lifecycle\n\n")
	b.WriteString("```text\n")
	b.WriteString("RunStarted\n")
	b.WriteString("   -> TextDelta / ReasoningDelta / ToolCallStarted / ToolCallFinished\n")
	b.WriteString("   -> FileChanged / CommandStarted / CommandFinished / StatusUpdate\n")
	b.WriteString("   -> UsageDelta / LogLine\n")
	b.WriteString("   -> InputRequested -> ProvideInput -> continue\n")
	b.WriteString("   -> ApprovalRequested -> ResolveApproval -> continue\n")
	b.WriteString("   -> RunReportedDone\n")
	b.WriteString("```\n")
	b.WriteString("\n`RunReportedDone` only records provider self-reported completion.\n")
	b.WriteString("Task completion still depends on validation after the terminal event.\n")
}

func writeAssertions(b *bytes.Buffer, assertions []string) {
	b.WriteString("\n## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- `%s`\n", assertion)
	}
}
