package main

import "strings"

func renderDetail(d detailDoc) string {
	return renderDoc(d.Title, "../native-config-overlay.md", func(b *strings.Builder) {
		for _, block := range d.Blocks {
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
	case "ordered":
		writeOrdered(b, block.Items)
	case "table":
		writeTable(b, block)
	case "code":
		b.WriteString("```" + block.Language + "\n" + block.Code + "\n```\n\n")
	}
}
