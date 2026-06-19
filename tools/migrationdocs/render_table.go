package main

import (
	"fmt"
	"strings"
)

func writeTable(b *strings.Builder, block block) {
	b.WriteString("| " + strings.Join(block.Columns, " | ") + " |\n")
	b.WriteString("|")
	for range block.Columns {
		b.WriteString(" --- |")
	}
	b.WriteByte('\n')
	for _, row := range block.Rows {
		fmt.Fprintf(b, "| %s |\n", strings.Join(row, " | "))
	}
	b.WriteByte('\n')
}
