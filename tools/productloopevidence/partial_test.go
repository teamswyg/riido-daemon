package main

import (
	"testing"
	"time"
)

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
