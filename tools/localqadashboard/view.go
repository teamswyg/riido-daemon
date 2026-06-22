package main

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
