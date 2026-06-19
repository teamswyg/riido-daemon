package main

import (
	"fmt"
	"strings"
)

func renderBoundaryCriteria(m manifest) string {
	return renderDoc("Boundary Criteria", func(b *strings.Builder) {
		writeBullets(b, []string{
			"Figma is product/design evidence, not daemon durable SSOT.",
			fmt.Sprintf("%s owns canonical UI coverage.", m.SourceCoverageManifestProvenance.Repo),
			m.BoundaryPolicy.Summary,
			"daemon consumes accepted assignments, runtime/model/instruction snapshots, provider detection/liveness, and stop/cancel/lifecycle commands.",
			"daemon does not own client copy, sorting, dropdown, modal, scroll, animation, timestamp, fixture row, workspace selection, waitlist, or marketing consent.",
		})
	})
}

func renderChangeLoop(m manifest) string {
	return renderDoc("Change Loop", func(b *strings.Builder) {
		b.WriteString("Top-down:\n\n")
		writeBullets(b, []string{m.BoundaryPolicy.TopDown})
		b.WriteString("Bottom-up:\n\n")
		writeBullets(b, []string{m.BoundaryPolicy.BottomUp})
	})
}
