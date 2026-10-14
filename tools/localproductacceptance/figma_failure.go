package main

func failedFigmaIntentScenarios(path string, err error) []scenario {
	repair := &repair{
		Class:   "figma_intent_manifest_unavailable",
		Owner:   "local-qa",
		Mode:    "manual",
		Summary: "Provide the Figma daemon-boundary intent manifest for local QA.",
	}
	return []scenario{
		{
			ID:             "figma.intent.catalog",
			Status:         statusFailed,
			Endpoint:       path,
			FailureSummary: err.Error(),
			Repair:         repair,
		},
		{
			ID:             "figma.onboarding",
			Status:         statusFailed,
			FailureSummary: err.Error(),
			Repair:         repair,
		},
		{
			ID:             "figma.runtime.settings",
			Status:         statusFailed,
			FailureSummary: err.Error(),
			Repair:         repair,
		},
		{
			ID:             "figma.runtime.detail",
			Status:         statusFailed,
			FailureSummary: err.Error(),
			Repair:         repair,
		},
	}
}
