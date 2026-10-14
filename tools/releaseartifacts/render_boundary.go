package main

import "strings"

func renderReviewBoundary(m manifest) string {
	return renderDoc("Review Boundary", "../release-artifacts.md", func(b *strings.Builder) {
		b.WriteString("The release archive must not include:\n\n")
		writeBullets(b, m.ForbiddenArchiveItems)
		b.WriteString("These constraints are inherited from:\n\n")
		writeBullets(b, m.InheritedRefs)
	})
}
