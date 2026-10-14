package main

import "testing"

func TestBuildDeploymentGateBlocksPartialCoverage(t *testing.T) {
	evidence := runEvidence{
		Status:         statusPassed,
		CoverageStatus: statusPartial,
		OpenRepairs: []runRepair{{
			Class:   "provider_auth_required",
			Owner:   "human",
			Mode:    "manual",
			Summary: "login required",
		}},
		ClosedLoops: []runLoopCandidate{{ID: "close-x"}},
	}
	got := buildDeploymentGate(evidence)
	if got.Status != deploymentGateBlocked {
		t.Fatalf("status=%q", got.Status)
	}
	if got.ObservedCoverageStatus != statusPartial || len(got.Blockers) != 3 {
		t.Fatalf("gate=%+v", got)
	}
}

func TestBuildDeploymentGateReadyForPassedCoverage(t *testing.T) {
	evidence := runEvidence{
		Status:         statusPassed,
		CoverageStatus: statusPassed,
	}
	got := buildDeploymentGate(evidence)
	if got.Status != deploymentGateReady || len(got.Blockers) != 0 {
		t.Fatalf("gate=%+v", got)
	}
	if got.StrictCommand == "" || got.RequiredCoverageStatus != statusPassed {
		t.Fatalf("gate=%+v", got)
	}
}
