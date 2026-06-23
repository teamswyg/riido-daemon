package main

func buildCoverage(
	m coverageManifest,
	e providerEvidenceFile,
	x externalEvidenceFile,
) ([]coverageRow, coverageSummary) {
	providers := providerIndex(e)
	externals := externalIndex(x.Scenarios)
	rows := make([]coverageRow, 0, len(m.Scenarios))
	for _, scenario := range m.Scenarios {
		row := coverageRow{
			ID:      scenario.ID,
			Title:   scenario.Title,
			Tier:    scenario.Tier,
			Surface: scenario.Surface,
			Status:  "not_verified",
		}
		if scenario.Evidence == "provider" {
			row = providerCoverageRow(row, providers[scenario.ProviderID])
		}
		if scenario.Evidence == "external" {
			row = externalCoverageRow(row, externals[scenario.ID])
		}
		rows = append(rows, row)
	}
	return rows, summarizeCoverage(rows)
}

func externalIndex(scenarios []externalScenario) map[string]externalScenario {
	out := map[string]externalScenario{}
	for _, scenario := range scenarios {
		out[scenario.ID] = scenario
	}
	return out
}

func providerIndex(e providerEvidenceFile) map[string]providerEvidence {
	out := map[string]providerEvidence{}
	for _, provider := range e.Providers {
		if provider.EvidenceArtifact == "" {
			provider.EvidenceArtifact = e.EvidenceArtifact
		}
		if provider.ExpiresAt == "" {
			provider.ExpiresAt = e.ExpiresAt
		}
		out[provider.ID] = provider
	}
	return out
}

func providerCoverageRow(row coverageRow, provider providerEvidence) coverageRow {
	if provider.ID == "" {
		return row
	}
	row.Status = provider.IntegrationStatus
	row.Evidence = provider.EvidenceArtifact
	row.ExpiresAt = provider.ExpiresAt
	if provider.Repair != nil {
		row.Repair = *provider.Repair
	}
	return row
}
