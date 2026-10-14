package main

import (
	"fmt"
	"strings"
)

func renderFigmaDaemon(m manifest) string {
	return renderDoc("Figma Daemon Boundaries", "../context-map.md", func(b *strings.Builder) {
		renderFigmaSections(b, m.FigmaDaemon.Sections)
		fmt.Fprintf(b, "Direct host-helper actions: %s\n\n", m.FigmaDaemon.DirectHostHelperRule)
		fmt.Fprintf(b, "Assignment authorization: %s\n", m.FigmaDaemon.AssignmentAuthorizationRule)
	})
}

func renderFigmaOnboarding(m manifest) string {
	return renderDoc("Figma Onboarding Boundaries", "../context-map.md", func(b *strings.Builder) {
		renderFigmaSections(b, m.FigmaOnboarding.Sections)
		b.WriteString("The daemon must not hard-code:\n\n")
		writeBullets(b, m.FigmaOnboarding.MustNotHardcode)
	})
}

func renderFigmaSections(b *strings.Builder, rows []figmaSection) {
	for _, row := range rows {
		fmt.Fprintf(b, "## %s\n\n", inlineCodeList(row.Refs))
		fmt.Fprintf(b, "Daemon scope: %s\n\n", row.DaemonScope)
		b.WriteString("Not owned:\n\n")
		writeBullets(b, row.NotOwned)
		b.WriteByte('\n')
	}
}
