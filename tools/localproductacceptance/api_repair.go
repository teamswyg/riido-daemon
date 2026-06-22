package main

import "strings"

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

func apiRepairForPayload(payload map[string]any) *repair {
	if strings.Contains(payloadErrorSummary(payload), "runtime binding") {
		return runtimeBindingRepair()
	}
	return apiRuntimeRepair()
}

func runtimeBindingRepair() *repair {
	return &repair{
		Class:            "ai_agent_runtime_binding_required",
		Owner:            "control-plane/daemon",
		Mode:             "manual",
		Summary:          "Runtime snapshot/binding is missing; start the daemon and verify agent runtime registration.",
		SuggestedCommand: "riido daemon stop --force || true; riido daemon start",
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
