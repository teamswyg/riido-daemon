package main

import (
	"fmt"
	"strings"
)

func renderPage(p page) string {
	return renderDoc(p.Title, func(b *strings.Builder) {
		if p.BackPath != "" {
			fmt.Fprintf(b, "[Back to %s](%s)\n\n", p.BackTitle, p.BackPath)
		}
		for _, block := range p.Blocks {
			renderBlock(b, block)
		}
	})
}

func renderBlock(b *strings.Builder, block block) {
	switch block.Kind {
	case "heading":
		b.WriteString("## " + block.Text + "\n\n")
	case "paragraph":
		b.WriteString(block.Text + "\n\n")
	case "bullets":
		writeBullets(b, block.Items)
		b.WriteByte('\n')
	case "links":
		writeLinks(b, block.Links)
		b.WriteByte('\n')
	case "table":
		writeTable(b, block)
	case "code":
		b.WriteString("```" + block.Language + "\n" + block.Code + "\n```\n\n")
	}
}
