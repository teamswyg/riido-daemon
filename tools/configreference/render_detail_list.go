package main

import (
	"fmt"
	"strings"
)

func renderDetailBullets(b *strings.Builder, items []string) {
	for _, item := range items {
		fmt.Fprintf(b, "- %s\n", item)
	}
	b.WriteByte('\n')
}

func renderOrderedList(b *strings.Builder, items []string) {
	for i, item := range items {
		fmt.Fprintf(b, "%d. %s\n", i+1, item)
	}
	b.WriteByte('\n')
}

func renderTable(b *strings.Builder, table *DetailTable) {
	if table == nil {
		return
	}
	fmt.Fprintf(b, "| %s |\n", strings.Join(table.Headers, " | "))
	separators := make([]string, 0, len(table.Headers))
	for range table.Headers {
		separators = append(separators, "---")
	}
	fmt.Fprintf(b, "| %s |\n", strings.Join(separators, " | "))
	for _, row := range table.Rows {
		fmt.Fprintf(b, "| %s |\n", strings.Join(row, " | "))
	}
	b.WriteByte('\n')
}
