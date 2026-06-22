package main

func assignmentBlockedScenarios(run taskAssignmentRun) []scenario {
	summary := "Need two assignment creates before multi-assignment, SSE, and thread message checks."
	repair := dominantAssignmentRepair(run)
	return []scenario{
		blockedTaskScenario("contract.task.multi_assignment", summary, repair),
		blockedTaskScenario("contract.task.thread_subscription", summary, repair),
		blockedTaskScenario("contract.task.sse_replay", summary, repair),
		blockedTaskScenario("contract.task.thread_message", summary, repair),
	}
}

func dominantAssignmentRepair(run taskAssignmentRun) *repair {
	for _, sc := range run.Scenarios {
		if sc.Repair != nil && sc.Repair.Class == "ai_agent_runtime_binding_required" {
			return runtimeBindingRepair()
		}
	}
	return apiRuntimeRepair()
}

func blockedTaskScenario(id, summary string, repair *repair) scenario {
	return scenario{
		ID:             id,
		Status:         statusSkipped,
		FailureSummary: summary,
		Repair:         repair,
	}
}
