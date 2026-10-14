package main

func externalCoverageRow(row coverageRow, scenario externalScenario) coverageRow {
	if scenario.ID == "" {
		return row
	}
	row.Status = scenario.Status
	row.Detail = scenario.FailureSummary
	row.Evidence = scenario.Evidence
	row.ExpiresAt = scenario.ExpiresAt
	if scenario.Screenshot != "" {
		row.Screenshot = screenshotHref(scenario.Screenshot)
	}
	if scenario.Repair != nil {
		row.Repair = *scenario.Repair
	}
	return row
}

func screenshotHref(path string) string {
	const prefix = ".riido-local/screenshots/"
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		return "screenshots/" + path[len(prefix):]
	}
	return path
}
