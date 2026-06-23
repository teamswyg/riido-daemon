package main

const (
	deploymentGateReady   = "ready"
	deploymentGateBlocked = "blocked"
)

func deploymentGateStatus(run localRunEvidence) string {
	if run.DeploymentGate.Status != "" {
		return run.DeploymentGate.Status
	}
	if run.CoverageStatus == statusPassed {
		return deploymentGateReady
	}
	return deploymentGateBlocked
}
