package main

import (
	"bytes"
	"fmt"
)

func writeAbsentSurfaces(b *bytes.Buffer, surfaces []AbsentSurface) {
	if len(surfaces) == 0 {
		return
	}
	b.WriteString("## Forbidden Claims\n\n| Claim | Reason |\n| --- | --- |\n")
	for _, surface := range surfaces {
		fmt.Fprintf(b, "| `%s` | %s |\n", surface.Name, surface.Reason)
	}
	b.WriteString("\n")
}
