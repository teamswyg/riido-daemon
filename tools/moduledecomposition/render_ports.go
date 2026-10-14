package main

import (
	"fmt"
	"strings"
)

func renderHexagonalPorts(m manifest) string {
	return renderDoc("Hexagonal Ports", "../module-decomposition.md", func(b *strings.Builder) {
		b.WriteString("| Port | Package | Adapters |\n| --- | --- | --- |\n")
		for _, port := range m.Ports {
			fmt.Fprintf(b, "| %s | `%s` | %s |\n", port.Port, port.Package, port.Adapters)
		}
		b.WriteString("\nAdapters translate into provider-neutral types before crossing inward.\n")
		b.WriteString("Raw provider payloads may be retained only behind event/adapter audit contracts.\n")
	})
}
