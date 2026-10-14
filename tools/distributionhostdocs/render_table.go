package main

import "strings"

func writeTable(b *strings.Builder, block block) {
	b.WriteString("| " + strings.Join(block.Columns, " | ") + " |\n")
	b.WriteString("|")
	for range block.Columns {
		b.WriteString(" --- |")
	}
	b.WriteByte('\n')
	for _, row := range block.Rows {
		b.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}
	b.WriteByte('\n')
}
