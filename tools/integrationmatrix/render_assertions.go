package main

import (
	"fmt"
	"strings"
)

func renderAssertions(m manifest) string {
	return renderDoc("Integration Assertions", "../integration-matrix.md", func(b *strings.Builder) {
		b.WriteString("| Provider | Integration assertion |\n| --- | --- |\n")
		for _, provider := range m.ProviderValidation.Providers {
			fmt.Fprintf(b, "| %s | %s |\n", provider.DisplayName, assertionText(provider))
		}
		b.WriteString("\nMatrix-level assertions:\n\n")
		writeBullets(b, m.Assertions)
	})
}

func assertionText(provider providerEvidence) string {
	parts := append([]string{}, provider.PassEvidence...)
	parts = append(parts, provider.NegativeOrLimitedEvidence...)
	parts = append(parts, provider.RequiredSchedulingGate...)
	return strings.Join(parts, "; ")
}
