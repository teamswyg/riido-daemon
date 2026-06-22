package main

func buildCoverage(m coverageManifest, e providerEvidenceFile) ([]coverageRow, coverageSummary) {
	providers := providerIndex(e.Providers)
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
		rows = append(rows, row)
	}
	return rows, summarizeCoverage(rows)
}

func providerIndex(providers []providerEvidence) map[string]providerEvidence {
	out := map[string]providerEvidence{}
	for _, provider := range providers {
		out[provider.ID] = provider
	}
	return out
}

func providerCoverageRow(row coverageRow, provider providerEvidence) coverageRow {
	if provider.ID == "" {
		return row
	}
	row.Status = provider.IntegrationStatus
	if provider.Repair != nil {
		row.Repair = *provider.Repair
	}
	return row
}
