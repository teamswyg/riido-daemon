package main

func summarizeAssignableAgents(payload map[string]any) map[string]any {
	return map[string]any{"agents_count": arrayLen(payload["agents"])}
}

func summarizeAssignedProfiles(payload map[string]any) map[string]any {
	profiles, _ := payload["assigned_agent_profiles"].(map[string]any)
	return map[string]any{
		"workspace_id_present":     stringPresent(payload["workspace_id"]),
		"assigned_task_keys_count": len(profiles),
	}
}

func summarizeTaskAction(payload map[string]any) map[string]any {
	return map[string]any{
		"agent_id":          payload["agent_id"],
		"assignment_id":     payload["assignment_id"],
		"thread_id":         payload["thread_id"],
		"run_id":            payload["run_id"],
		"work_status":       payload["work_status"],
		"assignment_state":  payload["assignment_state"],
		"active_stream_set": payload["active_stream"] != nil,
	}
}

func summarizeSubscription(payload map[string]any) map[string]any {
	stream, _ := payload["stream"].(map[string]any)
	return map[string]any{
		"active_thread_filters_count": arrayLen(payload["active_thread_filters"]),
		"stream_href_present":         stringPresent(stream["href"]),
		"stream_event_type":           stream["event_type"],
	}
}

func distinctAssignmentScenario(plan taskMutationPlan, first, second scenario) scenario {
	out := scenario{ID: "contract.task.multi_assignment", Status: statusPassed}
	firstAssignment, _ := first.Observed["assignment_id"].(string)
	secondAssignment, _ := second.Observed["assignment_id"].(string)
	firstThread, _ := first.Observed["thread_id"].(string)
	secondThread, _ := second.Observed["thread_id"].(string)
	out.Observed = map[string]any{
		"assignment_ids_distinct": firstAssignment != "" && firstAssignment != secondAssignment,
		"thread_ids_distinct":     firstThread != "" && firstThread != secondThread,
		"task_id_source":          plan.TaskIDSource,
		"same_runtime_kind_pair":  plan.Pair.First.RuntimeKind != "" && plan.Pair.First.RuntimeKind == plan.Pair.Second.RuntimeKind,
		"first_runtime_kind":      plan.Pair.First.RuntimeKind,
		"second_runtime_kind":     plan.Pair.Second.RuntimeKind,
	}
	if first.Status != statusPassed || second.Status != statusPassed {
		out.Status = statusFailed
		out.FailureSummary = "one or more assignment creates failed"
		out.Repair = apiRuntimeRepair()
	}
	if out.Observed["assignment_ids_distinct"] != true || out.Observed["thread_ids_distinct"] != true {
		out.Status = statusFailed
		out.FailureSummary = "multi assignment collapsed by assignment_id or thread_id"
	}
	return out
}
