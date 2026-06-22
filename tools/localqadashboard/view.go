package main

type dashboardView struct {
	Evidence        providerEvidenceFile
	Run             localRunEvidence
	CoverageRows    []coverageRow
	CoverageSummary coverageSummary
}
