package main

import "strings"

func matchingFigmaEntries(entries []figmaIntentEntry, needle string) []figmaIntentEntry {
	var out []figmaIntentEntry
	for _, entry := range entries {
		if strings.Contains(entry.Name, needle) {
			out = append(out, entry)
		}
	}
	return out
}
