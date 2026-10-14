package main

import (
	"fmt"
	"strings"
)

func renderAnchors(b *strings.Builder, anchors []Anchor) {
	b.WriteString("## Coverage Anchors\n\n")
	for _, anchor := range anchors {
		fmt.Fprintf(b, "- `%s`: %s (`%s`)\n", anchor.Name, anchor.Meaning, anchor.File)
	}
	b.WriteByte('\n')
}

func renderAssertions(b *strings.Builder, assertions []string) {
	b.WriteString("## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- %s\n", assertion)
	}
}
