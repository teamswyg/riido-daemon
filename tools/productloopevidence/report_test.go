package main

import (
	"testing"
	"time"
)

func TestBuildProductAcceptanceFindsMissingSignals(t *testing.T) {
	m := manifest{OutcomeSignals: []outcomeSignal{
		{ID: "assignment_completion", ScenarioIDs: []string{"contract.task.multi_assignment"}},
		{ID: "missing", ScenarioIDs: []string{"not.present"}},
	}}
	local := localAcceptanceSource{Scenarios: []coverageScenario{{ID: "contract.task.multi_assignment"}}}
	got := buildProductAcceptance(m, local)
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

func TestCollectCandidatesPromotesPartialEvidence(t *testing.T) {
	meta := metaComplexity{Status: statusPartial, PartialReason: "entrypoint budget"}
	product := productAcceptance{Status: statusPartial, MissingSignalIDs: []string{"time_to_first_event"}}
	partial := partialReduction{
		Status:                    statusPartial,
		CandidateAgeUnknownCount:  1,
		LocalQARunEvidencePresent: true,
	}
	got := collectCandidates(meta, product, partial)
	if len(got) != 3 {
		t.Fatalf("candidates = %+v", got)
	}
	for _, candidate := range got {
		if len(candidate.RequiredNextArtifacts) == 0 {
			t.Fatalf("candidate missing next artifacts: %+v", candidate)
		}
	}
}

func TestBuildPartialReductionComputesCandidateAge(t *testing.T) {
	now := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)
	m := manifest{Thresholds: thresholds{StalePartialAfterDays: 7}}
	reg := registrySource{Loops: []registryLoop{
		{ID: "fresh", Kind: "systematization-audit", CandidateCreatedAt: "2026-06-24", PromotionTarget: "verifier"},
		{ID: "stale", Kind: "systematization-audit", CandidateCreatedAt: "2026-06-01", PromotionTarget: "gate"},
		{ID: "unknown", Kind: "systematization-audit"},
		{ID: "closed", Kind: "closed-loop"},
	}}
	got := buildPartialReductionAt(t.TempDir(), m, reg, qaSystemSource{}, now)
	if got.CandidateAgeUnknownCount != 1 || got.StaleCandidateCount != 1 {
		t.Fatalf("partial reduction = %+v", got)
	}
	if len(got.CandidateAges) != 2 || got.CandidateAges[0].AgeDays != 2 || !got.CandidateAges[1].Stale {
		t.Fatalf("candidate ages = %+v", got.CandidateAges)
	}
}
