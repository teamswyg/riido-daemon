package main

import (
	"fmt"
	"strings"
)

func renderIndex(m manifest) string {
	return renderDoc(m.InvariantsIndex.Title, func(b *strings.Builder) {
		b.WriteString("[Back to runtime-scheduling.md](../runtime-scheduling.md)\n\n")
		for _, summary := range m.InvariantsIndex.Summary {
			fmt.Fprintf(b, "%s\n\n", summary)
		}
		b.WriteString("## Parts\n\n")
		writeLinks(b, m.InvariantsIndex.Parts)
	})
}
