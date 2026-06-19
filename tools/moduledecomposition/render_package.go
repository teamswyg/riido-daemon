package main

import (
	"fmt"
	"strings"
)

func renderPackageMap(m manifest) string {
	return renderDoc("Package Map", "../module-decomposition.md", func(b *strings.Builder) {
		b.WriteString("| Package | Role |\n| --- | --- |\n")
		for _, role := range m.PackageRoles {
			fmt.Fprintf(b, "| `%s` | %s |\n", role.Label, role.Role)
		}
		b.WriteString("\nPackage existence is verified with `go list -json ./...`.\n")
	})
}
