package main

func summarizeCoverage(rows []coverageRow) coverageSummary {
	summary := coverageSummary{Total: len(rows)}
	for _, row := range rows {
		switch row.Status {
		case "passed":
			summary.Passed++
		case "skipped":
			summary.Skipped++
		case "failed":
			summary.Failed++
		case "not_verified":
			summary.NotVerified++
		}
	}
	return summary
}
