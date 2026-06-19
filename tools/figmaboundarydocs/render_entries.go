package main

import (
	"fmt"
	"strings"
)

func renderScreenEntries(m manifest) string {
	return renderDoc("Screen Entries", func(b *strings.Builder) {
		b.WriteString("| Node | Screen | Daemon boundary |\n| --- | --- | --- |\n")
		for _, entry := range m.Entries {
			fmt.Fprintf(b, "| `%s` | %s | %s |\n", entry.NodeID, entry.Name, entry.DaemonScope)
		}
		b.WriteString("\nThe executable entry catalog is [`entries.riido.json`](entries.riido.json).\n")
	})
}

func renderFixtureVocabulary(m manifest) string {
	entry := findFixtureEntry(m)
	return renderDoc("Fixture Vocabulary", func(b *strings.Builder) {
		b.WriteString("Figma and older discussions used \"template\" loosely, but current SSOT has no daemon-owned agent template entity.\n\n")
		if entry.NodeID != "" {
			fmt.Fprintf(b, "`%s` is the fixture-selection entry. %s\n\n", entry.Name, entry.DaemonScope)
			fmt.Fprintf(b, "Client-owned fixture facts: %s.\n\n", sentenceList(entry.ClientOwnedFacts))
		}
		b.WriteString("Daemon never hardcodes fixture catalog, descriptions, or instruction copy; it only receives the assignment-time instruction/runtime/model snapshot.\n")
	})
}

func findFixtureEntry(m manifest) boundaryEntry {
	for _, entry := range m.Entries {
		if strings.Contains(strings.Join(entry.ClientOwnedFacts, " "), "fixture") {
			return entry
		}
	}
	return boundaryEntry{}
}
