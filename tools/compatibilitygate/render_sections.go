package main

import (
	"bytes"
	"fmt"
	"strings"
)

func writeOutputs(b *bytes.Buffer, outputs []string) {
	b.WriteString("\n## Outputs\n\n")
	for _, output := range outputs {
		fmt.Fprintf(b, "- %s\n", output)
	}
}

func writeFailureSemantics(b *bytes.Buffer, failures []FailureSemantic) {
	b.WriteString("\n## Failure Semantics\n\n| Case | Meaning | Evidence |\n| --- | --- | --- |\n")
	for _, failure := range failures {
		fmt.Fprintf(b, "| %s | %s | `%s` |\n", failure.Case, failure.Meaning, joinRefs(failure.SourceChecks))
	}
}

func writeAssertions(b *bytes.Buffer, assertions []string) {
	b.WriteString("\n## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- `%s`\n", assertion)
	}
}

func joinRefs(refs []string) string {
	return strings.Join(refs, "`, `")
}
