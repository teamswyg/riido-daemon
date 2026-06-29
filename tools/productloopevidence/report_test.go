package main

import "testing"

func TestBuildProductAcceptanceFindsMissingSignals(t *testing.T) {
	m := manifest{OutcomeSignals: []outcomeSignal{
		{ID: "assignment_completion", ScenarioIDs: []string{"contract.task.multi_assignment"}},
		{ID: "missing", ScenarioIDs: []string{"not.present"}},
	}}
	local := localAcceptanceSource{Scenarios: []coverageScenario{{ID: "contract.task.multi_assignment"}}}
	run := productRunOutcomeSource{
		State:          localQARunFresh,
		ScenarioStatus: map[string]string{"contract.task.multi_assignment": statusPassed},
	}
	got := buildProductAcceptance(m, local, run)
	if got.Status != statusPartial || got.LinkedSignalCount != 1 {
		t.Fatalf("acceptance = %+v", got)
	}
	if len(got.MissingSignalIDs) != 1 || got.MissingSignalIDs[0] != "missing" {
		t.Fatalf("missing signals = %+v", got.MissingSignalIDs)
	}
}

func TestBuildMappingCoverageRequiresVerifierClaims(t *testing.T) {
	reg := registrySource{BusinessClaims: []registryClaim{
		{ID: "covered", Files: []string{"a.go"}, Verifiers: []sourceCheck{{Name: "test", File: "a_test.go"}}},
		{ID: "open", Files: []string{"b.go"}},
	}}
	got := buildMappingCoverage(reg)
	if got.ClaimCount != 2 || got.ClaimWithVerifierCount != 1 {
		t.Fatalf("coverage = %+v", got)
	}
	if got.CoverageRatio != 0.5 {
		t.Fatalf("ratio = %v", got.CoverageRatio)
	}
}

func TestBuildMetaComplexityTreatsRoutedEntrypointsAsManaged(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "tools/a/main.go")
	writeFixture(t, root, "tools/b/main.go")
	m := manifest{Thresholds: thresholds{MaxEntrypointsBeforePartial: 1}}
	reg := registrySource{BusinessClaims: []registryClaim{
		{ID: "covered", Files: []string{"a.go"}, Verifiers: []sourceCheck{{Name: "test", File: "a_test.go"}}},
	}}
	routes := entrypointRouteMap{Routes: []entrypointRoute{{
		ID: "tools", Owner: "platform", Includes: []string{"tools/*/main.go"},
	}}}
	got := buildMetaComplexity(root, m, reg, routes)
	if got.Status != statusPassed || got.RouteCoverage.CoverageRatio != 1 {
		t.Fatalf("meta complexity = %+v", got)
	}
}
