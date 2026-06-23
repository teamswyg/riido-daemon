package main

func run(inputPath, externalPath, releasePath, runPath, schedulePath, infraPath string,
	manifestPath, outputPath, coveragePath string,
) error {
	evidence, err := loadProviderEvidence(inputPath)
	if err != nil {
		return err
	}
	external, err := loadExternalEvidence(externalPath)
	if err != nil {
		return err
	}
	external.Scenarios = append(external.Scenarios, runEvidenceScenarios(runPath)...)
	external.Scenarios = append(external.Scenarios, scheduleEvidenceScenarios(schedulePath)...)
	external.Scenarios = append(external.Scenarios, infraEvidenceScenarios(infraPath)...)
	external.Scenarios = append(external.Scenarios, releaseEvidenceScenarios(releasePath)...)
	runEvidence, _ := loadLocalRunEvidence(runPath)
	manifest, err := loadCoverageManifest(manifestPath)
	if err != nil {
		return err
	}
	rows, summary := buildCoverage(manifest, evidence, external)
	if err := writeCoverageSnapshot(coveragePath, rows, summary); err != nil {
		return err
	}
	rendered, err := renderDashboard(dashboardView{
		Evidence:        evidence,
		Run:             runEvidence,
		CoverageRows:    rows,
		CoverageSummary: summary,
	})
	if err != nil {
		return err
	}
	return writeText(outputPath, rendered)
}
