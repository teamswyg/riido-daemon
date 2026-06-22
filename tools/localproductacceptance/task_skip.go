package main

func skippedTaskScenario(in scenario, summary string) scenario {
	in.Status = statusSkipped
	in.FailureSummary = summary
	in.Repair = taskConfigRepair(summary)
	return in
}
