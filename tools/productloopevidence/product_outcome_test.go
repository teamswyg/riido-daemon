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

func TestProductLoopManifestRequiresClaudeApprovalRoundTrip(t *testing.T) {
	root := repoRoot()
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, signal := range m.OutcomeSignals {
		if signal.ID != "provider_approval_round_trip" {
			continue
		}
		found = true
		if len(signal.ScenarioIDs) != 1 ||
			signal.ScenarioIDs[0] != "provider.claude.web_approval_round_trip" {
			t.Fatalf("approval signal scenarios = %+v", signal.ScenarioIDs)
		}
	}
	if !found {
		t.Fatal("provider_approval_round_trip signal missing")
	}
}

func TestBuildProductAcceptanceShowsMissingOutcomeScenarioRows(t *testing.T) {
	m := manifest{OutcomeSignals: []outcomeSignal{{
		ID: "time_to_first_event",
		ScenarioIDs: []string{
			"contract.task.thread_subscription",
			"contract.task.sse_replay",
		},
	}}}
	local := localAcceptanceSource{Scenarios: []coverageScenario{
		{ID: "contract.task.thread_subscription"},
		{ID: "contract.task.sse_replay"},
	}}
	run := productRunOutcomeSource{
		State: localQARunFresh,
		ScenarioStatus: map[string]string{
			"contract.task.thread_subscription": "observed",
			"contract.task.sse_replay":          "not_verified",
		},
	}
	got := buildProductAcceptance(m, local, run)
	measure := got.MeasurementCandidates[0]
	if measure.OutcomeEvidenceLinked || len(measure.MissingOutcomeEvidenceScenarioIDs) != 1 {
		t.Fatalf("measure = %+v", measure)
	}
	if measure.MissingOutcomeEvidenceScenarioIDs[0] != "contract.task.sse_replay" {
		t.Fatalf("missing scenario rows = %+v", measure.MissingOutcomeEvidenceScenarioIDs)
	}
}
