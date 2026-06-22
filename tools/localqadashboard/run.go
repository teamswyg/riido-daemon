package main

func run(inputPath, manifestPath, outputPath string) error {
	evidence, err := loadProviderEvidence(inputPath)
	if err != nil {
		return err
	}
	manifest, err := loadCoverageManifest(manifestPath)
	if err != nil {
		return err
	}
	rows, summary := buildCoverage(manifest, evidence)
	rendered, err := renderDashboard(dashboardView{
		Evidence:        evidence,
		CoverageRows:    rows,
		CoverageSummary: summary,
	})
	if err != nil {
		return err
	}
	return writeText(outputPath, rendered)
}
