package main

import "testing"

func TestFigmaIntentScenariosUseGoldenReferences(t *testing.T) {
	scenarios := figmaIntentScenarios(
		"../../docs/30-architecture/figma-ai-agent-daemon-boundary/entries.riido.json",
		"../../docs/30-architecture/figma-ai-agent-daemon-boundary/golden/golden.riido.json",
		t.TempDir(),
	)
	byID := map[string]scenario{}
	for _, scenario := range scenarios {
		byID[scenario.ID] = scenario
	}
	if byID["figma.intent.catalog"].Status != statusPassed {
		t.Fatalf("catalog status = %q", byID["figma.intent.catalog"].Status)
	}
	onboarding := byID["figma.onboarding"]
	if onboarding.Status != statusPassed {
		t.Fatalf("onboarding status = %q", onboarding.Status)
	}
	if onboarding.Screenshot == "" || onboarding.Observed["golden"] == nil {
		t.Fatalf("onboarding golden evidence missing: %+v", onboarding)
	}
}
