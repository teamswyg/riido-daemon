package main

import "testing"

func TestStagingFixtureHandoffScenarioPassesCompleteFixture(t *testing.T) {
	cfg := domainJourneyTestConfig("cache.json")
	rows := completeDomainRows()
	got := stagingFixtureHandoffScenario(cfg, rows)
	if got.Status != statusPassed || got.Repair != nil {
		t.Fatalf("expected passed fixture handoff: %+v", got)
	}
}

func TestStagingFixtureHandoffScenarioRequiresInputs(t *testing.T) {
	cfg := domainJourneyTestConfig("cache.json")
	*cfg.apiToken, *cfg.workspaceID = "", ""
	got := stagingFixtureHandoffScenario(cfg, nil)
	if got.Status != statusPartial || got.Repair == nil {
		t.Fatalf("expected fixture repair evidence: %+v", got)
	}
	if got.Observed["replaces_inferred_id"] != "staging-fixture-handoff" {
		t.Fatalf("missing inferred replacement marker: %+v", got.Observed)
	}
}

func completeDomainRows() []scenario {
	rows := make([]scenario, 0, len(domainEntityDefs()))
	for _, entity := range domainEntityDefs() {
		rows = append(rows, scenario{ID: "domain.fixture." + entity.Key, Status: statusPassed})
	}
	return rows
}
