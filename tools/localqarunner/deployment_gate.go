package main

const (
	deploymentGateReady   = "ready"
	deploymentGateBlocked = "blocked"
	deploymentGateCommand = "go run ./tools/localqarunner -run-product -strict-coverage"
)

func applyDeploymentGate(evidence runEvidence) runEvidence {
	evidence.DeploymentGate = buildDeploymentGate(evidence)
	return evidence
}

func buildDeploymentGate(evidence runEvidence) runDeploymentGate {
	gate := runDeploymentGate{
		Status:                 deploymentGateReady,
		RequiredCoverageStatus: statusPassed,
		ObservedCoverageStatus: evidence.CoverageStatus,
		StrictCommand:          deploymentGateCommand,
	}
	gate.Blockers = appendRunStatusBlocker(gate.Blockers, evidence.Status)
	gate.Blockers = appendCoverageBlocker(gate.Blockers, evidence.CoverageStatus)
	gate.Blockers = appendRepairBlocker(gate.Blockers, len(evidence.OpenRepairs))
	if len(gate.Blockers) > 0 {
		gate.Status = deploymentGateBlocked
	}
	return gate
}

func appendRunStatusBlocker(blockers []runDeploymentGateBlocker, status string) []runDeploymentGateBlocker {
	if status == "" || status == statusPassed {
		return blockers
	}
	return append(blockers, runDeploymentGateBlocker{
		Code:    "run_status_not_passed",
		Summary: "local QA run did not pass",
		Count:   1,
	})
}

func appendCoverageBlocker(blockers []runDeploymentGateBlocker, status string) []runDeploymentGateBlocker {
	if status == "" || status == statusPassed {
		return blockers
	}
	return append(blockers, runDeploymentGateBlocker{
		Code:    "coverage_status_not_passed",
		Summary: "local QA coverage is not complete",
		Count:   1,
	})
}

func appendRepairBlocker(blockers []runDeploymentGateBlocker, count int) []runDeploymentGateBlocker {
	if count == 0 {
		return blockers
	}
	return append(blockers, runDeploymentGateBlocker{
		Code:    "open_repairs_present",
		Summary: "manual or automatic repairs are still open",
		Count:   count,
	})
}
