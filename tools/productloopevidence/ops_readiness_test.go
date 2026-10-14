package main

import "testing"

func TestProductLoopManifestDeclaresOperationalReadinessSignals(t *testing.T) {
	m, err := loadManifest(repoPath(repoRoot(), defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	signals := outcomeSignalSet(m)
	for _, id := range []string{
		"client_surface_anomaly_monitoring",
		"release_usability_regression",
		"network_resilience",
		"stress_capacity",
		"scale_chaos",
		"release_readiness",
	} {
		if !signals[id] {
			t.Fatalf("missing operational readiness signal %q", id)
		}
	}
}

func TestLocalAcceptanceManifestNamesOperationalReadinessScenarios(t *testing.T) {
	root := repoRoot()
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	var local localAcceptanceSource
	if err := loadJSON(repoPath(root, m.LocalAcceptanceManifest), &local); err != nil {
		t.Fatal(err)
	}
	scenarios := coverageScenarioSet(local)
	for _, id := range operationalScenarioIDs() {
		if !scenarios[id] {
			t.Fatalf("missing operational readiness scenario %q", id)
		}
	}
}

func outcomeSignalSet(m manifest) map[string]bool {
	out := make(map[string]bool, len(m.OutcomeSignals))
	for _, signal := range m.OutcomeSignals {
		out[signal.ID] = true
	}
	return out
}

func coverageScenarioSet(local localAcceptanceSource) map[string]bool {
	out := make(map[string]bool, len(local.Scenarios))
	for _, scenario := range local.Scenarios {
		out[scenario.ID] = true
	}
	return out
}
