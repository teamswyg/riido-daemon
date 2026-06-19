package main

import (
	"fmt"
	"strings"
)

func renderSplitRepo(m manifest) string {
	return renderDoc("Split-Repo Ownership", "../context-map.md", func(b *strings.Builder) {
		b.WriteString("Rules:\n\n")
		writeBullets(b, m.SplitRepo.Rules)
		b.WriteString("\nThe daemon must not redefine:\n\n")
		writeBullets(b, m.SplitRepo.DaemonMustNotRedefine)
		fmt.Fprintf(b, "\nThe daemon-side projection of Figma boundaries is [%s](%s).\n\n",
			m.SplitRepo.DownstreamBoundaryCheck, m.SplitRepo.DownstreamBoundaryCheck)
		fmt.Fprintln(b, m.SplitRepo.DownstreamBoundaryNote)
	})
}
