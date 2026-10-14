package main

import (
	"fmt"
	"strings"
)

func problemMapDoc(m model) string {
	return renderDoc("Problem Map", "Overview", "overview.md", func(b *strings.Builder) {
		b.WriteString("| ID | Symptom | Structural cause | Direction |\n| --- | --- | --- | --- |\n")
		for _, row := range m.Risk.Problems {
			fmt.Fprintf(b, "| %s | %s | %s | %s |\n", row.ID, row.Symptom, row.Cause, row.Direction)
		}
	})
}

func currentStructureDoc(m model) string {
	return renderDoc("Current Structure Evidence", "Overview", "overview.md", func(b *strings.Builder) {
		b.WriteString("| Observation | Current SSOT | Meaning |\n| --- | --- | --- |\n")
		for _, row := range m.Risk.StructureObservations {
			fmt.Fprintf(b, "| %s | `%s` | %s |\n", row.Observation, row.SSOT, row.Meaning)
		}
	})
}
