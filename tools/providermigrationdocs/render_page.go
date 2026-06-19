package main

import (
	"fmt"
	"strings"
)

func renderPage(p page) string {
	return renderDoc(p.Title, "../public-migration-status.md", func(b *strings.Builder) {
		writeParagraphs(b, p.Summary)
		if len(p.TableRows) > 0 {
			writeTable(b, p)
		}
		if len(p.Facts) > 0 {
			b.WriteString("Facts:\n\n")
			writeBullets(b, p.Facts)
			b.WriteByte('\n')
		}
		if len(p.Artifacts) > 0 {
			b.WriteString("Evidence artifacts:\n\n")
			writeBullets(b, p.Artifacts)
			b.WriteByte('\n')
		}
		if len(p.Links) > 0 {
			b.WriteString("Links:\n\n")
			for _, link := range p.Links {
				fmt.Fprintf(b, "- [%s](%s)\n", link.Title, link.Path)
			}
		}
	})
}

func writeTable(b *strings.Builder, p page) {
	b.WriteString("| " + strings.Join(p.TableColumns, " | ") + " |\n")
	b.WriteString("| " + strings.TrimRight(strings.Repeat("--- | ", len(p.TableColumns)), " ") + "\n")
	for _, row := range p.TableRows {
		b.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}
	b.WriteByte('\n')
}
