package main

import "testing"

func TestBuildDeploymentGateBlocksExpiredCoverageRows(t *testing.T) {
	evidence := runEvidence{
		Status:         statusPassed,
		CoverageStatus: statusPassed,
		Coverage: &runCoverage{Rows: []runCoverageRow{{
			ID:        "contract.task.thread_message",
			Status:    statusPassed,
			Evidence:  "coverage.json",
			ExpiresAt: "2000-01-01T00:00:00Z",
		}}},
	}
	got := buildDeploymentGate(evidence)
	if got.Status != deploymentGateBlocked {
		t.Fatalf("status=%q blockers=%+v", got.Status, got.Blockers)
	}
	if !hasGateBlocker(got.Blockers, "coverage_evidence_expired") {
		t.Fatalf("missing expired evidence blocker: %+v", got.Blockers)
	}
}

func TestBuildDeploymentGateAllowsFreshCoverageRows(t *testing.T) {
	evidence := runEvidence{
		Status:         statusPassed,
		CoverageStatus: statusPassed,
		Coverage: &runCoverage{Rows: []runCoverageRow{{
			ID:        "contract.task.thread_message",
			Status:    statusPassed,
			Evidence:  "coverage.json",
			ExpiresAt: "2999-01-01T00:00:00Z",
		}}},
	}
	got := buildDeploymentGate(evidence)
	if got.Status != deploymentGateReady {
		t.Fatalf("gate=%+v", got)
	}
}

func hasGateBlocker(blockers []runDeploymentGateBlocker, code string) bool {
	for _, blocker := range blockers {
		if blocker.Code == code {
			return true
		}
	}
	return false
}
