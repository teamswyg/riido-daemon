package main

type dashboardView struct {
	Evidence        providerEvidenceFile
	CoverageRows    []coverageRow
	CoverageSummary coverageSummary
}
