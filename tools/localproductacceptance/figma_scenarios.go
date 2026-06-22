package main

import "strings"

func figmaCatalogScenario(path string, entries []figmaIntentEntry) scenario {
	return scenario{
		ID:       "figma.intent.catalog",
		Status:   statusPassed,
		Endpoint: path,
		Observed: map[string]any{
			"entries_count": len(entries),
			"screen_names":  figmaScreenNames(entries),
		},
	}
}

func figmaScreenScenario(id string, entries []figmaIntentEntry, needle string) scenario {
	matches := matchingFigmaEntries(entries, needle)
	status := statusSkipped
	repairInfo := &repair{
		Class:   "figma_visual_golden_required",
		Owner:   "local-qa",
		Mode:    "manual",
		Summary: "Figma intent is loaded, but screenshot/golden visual comparison is not implemented yet.",
	}
	if len(matches) == 0 {
		status = statusFailed
		repairInfo = &repair{
			Class:   "figma_intent_missing",
			Owner:   "local-qa",
			Mode:    "manual",
			Summary: "Figma intent manifest does not contain the expected screen family.",
		}
	}
	return scenario{
		ID:       id,
		Status:   status,
		Observed: figmaObserved(matches),
		Repair:   repairInfo,
	}
}

func matchingFigmaEntries(entries []figmaIntentEntry, needle string) []figmaIntentEntry {
	var out []figmaIntentEntry
	for _, entry := range entries {
		if strings.Contains(entry.Name, needle) {
			out = append(out, entry)
		}
	}
	return out
}
