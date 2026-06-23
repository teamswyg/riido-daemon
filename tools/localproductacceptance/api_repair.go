package main

import "strings"

func apiConfigRepair(class, summary string) *repair {
	return &repair{Class: class, Owner: "local-qa", Mode: "manual", Summary: summary}
}

func apiRuntimeRepair() *repair {
	return &repair{
		Class:   "staging_api_unavailable_or_unauthorized",
		Owner:   "control-plane",
		Mode:    "manual",
		Summary: "Check token, workspace permission, and staging.ai-api.riido.io availability.",
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
		Summary:          "SaaS runtime snapshot/binding is missing; start a SaaS-connected daemon with device credentials and verify runtime registration.",
		SuggestedCommand: "riido daemon stop --force || true; RIIDO_SAAS_URL=https://staging.ai-api.riido.io RIIDO_DEVICE_ID=<device_id> RIIDO_DEVICE_SECRET=<device_secret> riido daemon start",
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
