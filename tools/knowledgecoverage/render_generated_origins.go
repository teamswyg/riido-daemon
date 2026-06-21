package main

import (
	"fmt"
	"strings"
)

func renderGeneratedOrigins(b *strings.Builder, origins []generatedOrigin) {
	b.WriteString("## Generated Reader Origins\n\n")
	if len(origins) == 0 {
		b.WriteString("_None._\n\n")
		return
	}
	b.WriteString("| Generator | Count | Sample paths |\n| --- | ---: | --- |\n")
	for _, origin := range origins {
		fmt.Fprintf(b, "| `%s` | %d | %s |\n", origin.Generator, origin.Count, samplePathText(origin.Samples))
	}
	b.WriteString("\n")
}
