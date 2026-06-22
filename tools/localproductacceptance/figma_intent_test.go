package main

import "testing"

func TestFigmaIntentScenariosExposeGoldenGap(t *testing.T) {
	scenarios := figmaIntentScenarios("../../docs/30-architecture/figma-ai-agent-daemon-boundary/entries.riido.json")
	byID := map[string]scenario{}
	for _, scenario := range scenarios {
		byID[scenario.ID] = scenario
	}
	if byID["figma.intent.catalog"].Status != statusPassed {
		t.Fatalf("catalog status = %q", byID["figma.intent.catalog"].Status)
	}
	onboarding := byID["figma.onboarding"]
	if onboarding.Status != statusSkipped {
		t.Fatalf("onboarding status = %q", onboarding.Status)
	}
	if onboarding.Repair == nil || onboarding.Repair.Class != "figma_visual_golden_required" {
		t.Fatalf("onboarding repair = %+v", onboarding.Repair)
	}
}
