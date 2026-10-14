package main

func applyStrictCoverage(evidence runEvidence) runEvidence {
	if evidence.StrictCoverage && strictCoverageBlocked(evidence) {
		evidence.Status = statusFailed
	}
	return evidence
}

func strictCoverageBlocked(evidence runEvidence) bool {
	if evidence.CoverageStatus != statusPassed {
		return true
	}
	return evidence.DeploymentGate.Status == deploymentGateBlocked
}
