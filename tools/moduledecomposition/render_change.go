package main

import "strings"

func renderChangeProcedure(m manifest) string {
	return renderDoc("Change Procedure", "../module-decomposition.md", func(b *strings.Builder) {
		writeBullets(b, m.ChangeProcedure)
		b.WriteString("Assertions:\n\n")
		writeBullets(b, m.Assertions)
	})
}
