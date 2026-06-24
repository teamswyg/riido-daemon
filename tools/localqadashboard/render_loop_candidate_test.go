package main

import (
	"strings"
	"testing"
)

func TestRenderDashboardIncludesClosedLoopCandidates(t *testing.T) {
	html, err := renderDashboard(dashboardView{
		Evidence: providerEvidenceFile{Status: "passed"},
		Run: localRunEvidence{
			Status:         "passed",
			CoverageStatus: "partial",
			ClosedLoops: []localRunLoopCandidate{{
				ID:           "close-x",
				Class:        "failed_probe",
				Reason:       "probe failed",
				NextEvidence: "add verifier",
				Graph:        localRunLoopCandidateGraph{NextLoop: "promote-x"},
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Closed-loop Candidates", "close-x", "promote-x"} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered dashboard missing %q", want)
		}
	}
}
