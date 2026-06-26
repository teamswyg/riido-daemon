package main

import "testing"

func TestBuildProductAcceptanceRequiresRunOutcomeEvidence(t *testing.T) {
	m := manifest{OutcomeSignals: []outcomeSignal{
		{ID: "assignment_completion", ScenarioIDs: []string{"contract.task.multi_assignment"}},
		{ID: "provider_recovery", ScenarioIDs: []string{"provider.claude"}},
	}}
	local := localAcceptanceSource{Scenarios: []coverageScenario{
		{ID: "contract.task.multi_assignment"},
		{ID: "provider.claude"},
	}}
	run := productRunOutcomeSource{
		State:          localQARunFresh,
		CoverageStatus: "partial",
		ScenarioStatus: map[string]string{
			"contract.task.multi_assignment": "not_verified",
			"provider.claude":                "observed",
		},
	}
	got := buildProductAcceptance(m, local, run)
	if got.Status != statusPartial || got.OutcomeEvidenceLinkedCount != 1 {
		t.Fatalf("acceptance = %+v", got)
	}
	if len(got.MissingOutcomeEvidenceSignalIDs) != 1 ||
		got.MissingOutcomeEvidenceSignalIDs[0] != "assignment_completion" {
		t.Fatalf("missing outcome evidence = %+v", got.MissingOutcomeEvidenceSignalIDs)
	}
}
