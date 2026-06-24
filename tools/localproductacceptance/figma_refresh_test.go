package main

import (
	"testing"
	"time"
)

func TestFigmaRefreshScenarioDetectsStaleGolden(t *testing.T) {
	observed := time.Date(2026, 6, 24, 14, 0, 0, 0, time.UTC)
	got := figmaRefreshScenario(
		observed,
		24*time.Hour,
		"../../docs/30-architecture/figma-ai-agent-daemon-boundary/entries.riido.json",
		"../../docs/30-architecture/figma-ai-agent-daemon-boundary/golden/golden.riido.json",
	)
	if got.Status != statusPartial || got.Repair == nil {
		t.Fatalf("expected stale golden repair evidence: %+v", got)
	}
	if got.Observed["replaces_inferred_id"] != "figma-refresh" || got.Observed["stale"] != true {
		t.Fatalf("unexpected observed proof: %+v", got.Observed)
	}
}

func TestFigmaRefreshScenarioPassesFreshGolden(t *testing.T) {
	observed := time.Date(2026, 6, 22, 14, 0, 0, 0, time.UTC)
	got := figmaRefreshScenario(
		observed,
		24*time.Hour,
		"../../docs/30-architecture/figma-ai-agent-daemon-boundary/entries.riido.json",
		"../../docs/30-architecture/figma-ai-agent-daemon-boundary/golden/golden.riido.json",
	)
	if got.Status != statusPassed || got.Repair != nil {
		t.Fatalf("expected fresh golden pass: %+v", got)
	}
}
