package main

import "testing"

func TestDeploymentGateStatusPrefersRunEvidence(t *testing.T) {
	run := localRunEvidence{
		CoverageStatus: statusPassed,
		DeploymentGate: localRunDeploymentGate{Status: deploymentGateBlocked},
	}
	if got := deploymentGateStatus(run); got != deploymentGateBlocked {
		t.Fatalf("status=%q", got)
	}
}

func TestDeploymentGateStatusFallsBackToCoverage(t *testing.T) {
	run := localRunEvidence{CoverageStatus: statusPassed}
	if got := deploymentGateStatus(run); got != deploymentGateReady {
		t.Fatalf("status=%q", got)
	}
}
