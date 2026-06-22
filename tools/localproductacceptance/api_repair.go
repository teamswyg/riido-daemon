package main

func apiConfigRepair(class, summary string) *repair {
	return &repair{Class: class, Owner: "local-qa", Mode: "manual", Summary: summary}
}

func apiRuntimeRepair() *repair {
	return &repair{
		Class:   "development_api_unavailable_or_unauthorized",
		Owner:   "control-plane",
		Mode:    "manual",
		Summary: "Check token, workspace permission, and development.ai-api.riido.io availability.",
	}
}

func taskConfigRepair(summary string) *repair {
	return &repair{Class: "task_flow_config_required", Owner: "local-qa", Mode: "manual", Summary: summary}
}

func failTaskScenario(id, summary string) scenario {
	return scenario{
		ID:             id,
		Status:         statusSkipped,
		FailureSummary: summary,
		Repair:         taskConfigRepair(summary),
	}
}
