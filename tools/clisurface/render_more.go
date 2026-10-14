package main

import (
	"fmt"
	"strings"
)

func renderProviders(b *strings.Builder, providers []string) {
	b.WriteString("## Provider Smoke Surface\n\n")
	fmt.Fprintf(b, "`riido bridge providers` must expose: %s.\n\n", codeList(providers))
}

func renderRelated(b *strings.Builder, sections []RelatedSection) {
	b.WriteString("## Detail Surfaces\n\n")
	for _, section := range sections {
		fmt.Fprintf(b, "- [%s](%s)\n", section.Title, section.Path)
	}
	b.WriteByte('\n')
}

func renderAssertions(b *strings.Builder, assertions []string) {
	b.WriteString("## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- %s\n", assertion)
	}
}

func commandGroupNames(groups []CommandGroup) []string {
	names := make([]string, 0, len(groups))
	for _, group := range groups {
		names = append(names, group.Name)
	}
	return names
}
