package main

func apiSkippedScenarios(repair *repair) []scenario {
	ids := []string{
		"contract.api.bootstrap",
		"contract.api.devices",
		"contract.api.profile_thumbnail.intent",
		"contract.task.assignable_agents",
		"contract.task.multi_assignment",
		"contract.task.thread_subscription",
		"contract.task.thread_message",
	}
	out := make([]scenario, 0, len(ids))
	for _, id := range ids {
		out = append(out, scenario{ID: id, Status: statusSkipped, Repair: repair})
	}
	return out
}
