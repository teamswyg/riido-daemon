package main

import (
	"fmt"
	"strings"
)

func writeOrdered(b *strings.Builder, items []string) {
	for i, item := range items {
		fmt.Fprintf(b, "%d. %s\n", i+1, item)
	}
	b.WriteByte('\n')
}

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
