package main

func figmaCatalogScenario(path string, entries []figmaIntentEntry) scenario {
	return scenario{
		ID:       "figma.intent.catalog",
		Status:   statusPassed,
		Endpoint: path,
		Observed: map[string]any{
			"entries_count": len(entries),
			"entries":       figmaObservedEntries(entries),
			"screen_names":  figmaScreenNames(entries),
		},
	}
}

func figmaScreenScenario(
	id string,
	entries []figmaIntentEntry,
	needle string,
	goldens map[string]figmaGoldenScreen,
	goldenErr error,
	screenshotDir string,
) scenario {
	matches := matchingFigmaEntries(entries, needle)
	if len(matches) == 0 {
		return figmaIntentMissingScenario(id, matches)
	}
	if goldenErr != nil {
		return figmaGoldenRequiredScenario(id, matches, goldenErr)
	}
	golden := goldens[id]
	screenshot, observed, err := verifyFigmaGolden(id, matches, golden, screenshotDir)
	if err != nil {
		return figmaGoldenStaleScenario(id, matches, err)
	}
	return scenario{
		ID:         id,
		Status:     statusPassed,
		Screenshot: screenshot,
		Observed:   observed,
	}
}

func figmaIntentScenario(id string, entries []figmaIntentEntry, needle string) scenario {
	matches := matchingFigmaEntries(entries, needle)
	if len(matches) == 0 {
		return figmaIntentMissingScenario(id, matches)
	}
	return scenario{
		ID:       id,
		Status:   statusPassed,
		Observed: figmaObserved(matches),
	}
}
