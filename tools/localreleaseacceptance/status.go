package main

func aggregateStatus(scenarios []scenario) string {
	for _, scenario := range scenarios {
		if scenario.Status != statusPassed {
			return statusFailed
		}
	}
	return statusPassed
}
