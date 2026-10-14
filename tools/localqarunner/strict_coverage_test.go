package main

import "testing"

func TestApplyStrictCoverageFailsPartialCoverage(t *testing.T) {
	evidence := runEvidence{
		Status:         statusPassed,
		CoverageStatus: statusPartial,
		StrictCoverage: true,
	}
	got := applyStrictCoverage(evidence)
	if got.Status != statusFailed {
		t.Fatalf("status=%q", got.Status)
	}
}

func TestApplyStrictCoverageKeepsDailyPartialPassing(t *testing.T) {
	evidence := runEvidence{
		Status:         statusPassed,
		CoverageStatus: statusPartial,
	}
	got := applyStrictCoverage(evidence)
	if got.Status != statusPassed {
		t.Fatalf("status=%q", got.Status)
	}
}

func TestApplyStrictCoverageFailsBlockedGate(t *testing.T) {
	evidence := runEvidence{
		Status:         statusPassed,
		CoverageStatus: statusPassed,
		StrictCoverage: true,
		DeploymentGate: runDeploymentGate{Status: deploymentGateBlocked},
	}
	got := applyStrictCoverage(evidence)
	if got.Status != statusFailed {
		t.Fatalf("status=%q", got.Status)
	}
}
