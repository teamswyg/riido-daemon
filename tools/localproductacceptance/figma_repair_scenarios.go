package main

func figmaIntentMissingScenario(id string, matches []figmaIntentEntry) scenario {
	return scenario{
		ID:       id,
		Status:   statusFailed,
		Observed: figmaObserved(matches),
		Repair: &repair{
			Class:   "figma_intent_missing",
			Owner:   "local-qa",
			Mode:    "manual",
			Summary: "Figma intent manifest does not contain the expected screen family.",
		},
	}
}

func figmaGoldenRequiredScenario(id string, matches []figmaIntentEntry, err error) scenario {
	return scenario{
		ID:             id,
		Status:         statusSkipped,
		FailureSummary: err.Error(),
		Observed:       figmaObserved(matches),
		Repair: &repair{
			Class:   "figma_visual_golden_required",
			Owner:   "local-qa",
			Mode:    "manual",
			Summary: "Figma intent is loaded, but visual golden reference is unavailable.",
		},
	}
}

func figmaGoldenStaleScenario(id string, matches []figmaIntentEntry, err error) scenario {
	return scenario{
		ID:             id,
		Status:         statusFailed,
		FailureSummary: err.Error(),
		Observed:       figmaObserved(matches),
		Repair: &repair{
			Class:   "figma_visual_golden_stale",
			Owner:   "local-qa",
			Mode:    "manual",
			Summary: "Refresh the Figma golden reference from the source node.",
		},
	}
}
