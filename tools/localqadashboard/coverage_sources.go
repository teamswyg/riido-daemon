package main

func withProviderSource(e providerEvidenceFile, path string) providerEvidenceFile {
	e.EvidenceArtifact = path
	for i := range e.Providers {
		e.Providers[i].EvidenceArtifact = path
		e.Providers[i].ExpiresAt = e.ExpiresAt
	}
	return e
}

func withExternalSource(e externalEvidenceFile, path string) externalEvidenceFile {
	e.Evidence = path
	e.Scenarios = withScenarioSource(e.Scenarios, path, e.ExpiresAt)
	return e
}

func withScenarioSource(rows []externalScenario, path, expires string) []externalScenario {
	out := make([]externalScenario, 0, len(rows))
	for _, row := range rows {
		if row.Evidence == "" {
			row.Evidence = path
		}
		if row.ExpiresAt == "" {
			row.ExpiresAt = expires
		}
		out = append(out, row)
	}
	return out
}

func withFallbackExpiry(rows []coverageRow, expires string) []coverageRow {
	for i := range rows {
		if rows[i].Evidence != "" && rows[i].ExpiresAt == "" {
			rows[i].ExpiresAt = expires
		}
	}
	return rows
}
