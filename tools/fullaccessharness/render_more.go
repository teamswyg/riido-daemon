package main

import (
	"fmt"
	"strings"
)

func renderFacts(b *strings.Builder, manifest Manifest) {
	fmt.Fprintln(b, "## Implemented Facts")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Fact | Summary | Evidence |")
	fmt.Fprintln(b, "| --- | --- | --- |")
	for _, fact := range manifest.Facts {
		fmt.Fprintf(b, "| %s | %s | `%s` |\n", fact.Name, fact.Summary, strings.Join(fact.SourceChecks, "`, `"))
	}
	fmt.Fprintln(b)
}

func renderAbsent(b *strings.Builder, manifest Manifest) {
	fmt.Fprintln(b, "## Absent Surfaces")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Surface | Scope | Reason |")
	fmt.Fprintln(b, "| --- | --- | --- |")
	for _, surface := range manifest.AbsentSurfaces {
		fmt.Fprintf(b, "| %s | `%s` | %s |\n", surface.Name, strings.Join(surface.Scope, "`, `"), surface.Reason)
	}
	fmt.Fprintln(b)
}

func renderAssertions(b *strings.Builder, manifest Manifest) {
	fmt.Fprintln(b, "## Assertions")
	fmt.Fprintln(b)
	for _, assertion := range manifest.Assertions {
		fmt.Fprintf(b, "- %s\n", assertion)
	}
}
