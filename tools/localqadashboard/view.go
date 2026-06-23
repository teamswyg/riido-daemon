package main

import "time"

type dashboardView struct {
	Evidence        providerEvidenceFile
	Run             localRunEvidence
	CoverageRows    []coverageRow
	CoverageSummary coverageSummary
}

func (view dashboardView) ObservedAt() string {
	if view.Run.ObservedAt != "" {
		return view.Run.ObservedAt
	}
	return view.Evidence.ObservedAt
}

func (view dashboardView) ExpiresAt() string {
	if view.Run.ExpiresAt != "" {
		return view.Run.ExpiresAt
	}
	return view.Evidence.ExpiresAt
}

func (view dashboardView) FreshnessStatus() string {
	expires, err := time.Parse(time.RFC3339, view.ExpiresAt())
	if err != nil {
		return "unknown"
	}
	if time.Now().UTC().Before(expires) {
		return "fresh"
	}
	return "expired"
}

func (view dashboardView) DeploymentGateStatus() string {
	if view.Run.CoverageStatus == statusPassed {
		return "ready"
	}
	return "blocked"
}
