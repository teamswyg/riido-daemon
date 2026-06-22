package main

func taskSkipped(includeAssignable bool, summary string) []scenario {
	out := []scenario{
		failTaskScenario("contract.task.multi_assignment", summary),
		failTaskScenario("contract.task.thread_subscription", summary),
		failTaskScenario("contract.task.sse_replay", summary),
		failTaskScenario("contract.task.thread_message", summary),
	}
	if includeAssignable {
		assignable := failTaskScenario("contract.task.assignable_agents", summary)
		out = append([]scenario{assignable}, out...)
	}
	return out
}
