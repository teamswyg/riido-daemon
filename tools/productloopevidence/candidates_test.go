package main

import "testing"

func TestCollectCandidatesPromotesPartialEvidence(t *testing.T) {
	meta := metaComplexity{Status: statusPartial, PartialReason: "entrypoint budget"}
	product := productAcceptance{Status: statusPartial, MissingSignalIDs: []string{"time_to_first_event"}}
	partial := partialReduction{
		Status:                    statusPartial,
		CandidateAgeUnknownCount:  1,
		LocalQARunEvidencePresent: true,
	}
	got := collectCandidates(meta, product, qaScheduleEvidence{}, partial)
	if len(got) != 3 {
		t.Fatalf("candidates = %+v", got)
	}
	for _, candidate := range got {
		if len(candidate.RequiredNextArtifacts) == 0 {
			t.Fatalf("candidate missing next artifacts: %+v", candidate)
		}
		if candidate.Graph.Observation == "" || candidate.Graph.NextLoop == "" {
			t.Fatalf("candidate missing graph: %+v", candidate)
		}
	}
}

func TestCollectCandidatesShowsOpenCandidateDebt(t *testing.T) {
	partial := partialReduction{
		CandidateCount:            2,
		LocalQARunEvidenceFresh:   true,
		LocalQARunEvidenceState:   localQARunFresh,
		ClosedLoopCandidateIDs:    []string{"a", "b"},
		CandidateAgeUnknownCount:  0,
		StaleCandidateCount:       0,
		LocalQARunEvidencePresent: true,
	}
	got := collectCandidates(metaComplexity{}, productAcceptance{}, qaScheduleEvidence{}, partial)
	if len(got) != 1 || got[0].ID != "partial-reduction-open-candidate-debt" {
		t.Fatalf("candidates = %+v", got)
	}
	if got[0].Graph.NextLoop != "local-qa-candidate-decision" {
		t.Fatalf("candidate graph = %+v", got[0].Graph)
	}
}
