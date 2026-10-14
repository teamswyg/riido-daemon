package main

import (
	"fmt"
	"strings"
)

func renderChangeProcedure(m manifest) string {
	return renderDoc("Change Procedure", "../context-map.md", func(b *strings.Builder) {
		fmt.Fprintf(b, "%s\nThe same PR must update:\n\n", m.ChangeProcedure.Summary)
		writeBullets(b, m.ChangeProcedure.SamePRUpdates)
	})
}
