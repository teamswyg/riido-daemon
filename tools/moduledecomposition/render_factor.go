package main

import (
	"fmt"
	"strings"
)

func renderFactorBoundary(m manifest) string {
	return renderDoc("12-Factor Boundary", "../module-decomposition.md", func(b *strings.Builder) {
		fmt.Fprintf(b, "Configuration is injected through %s.\n\n", m.FactorBoundary.Configuration)
		fmt.Fprintf(b, "Test-only gates use `%s`.\n\n", m.FactorBoundary.TestGates)
		fmt.Fprintf(b, "%s\n\n", m.FactorBoundary.State)
		fmt.Fprintf(b, "%s\n", m.FactorBoundary.Listener)
	})
}
