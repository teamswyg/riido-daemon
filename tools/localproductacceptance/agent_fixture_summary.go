package main

func summarizeAgentCreate(payload map[string]any) map[string]any {
	agent, _ := payload["agent"].(map[string]any)
	return map[string]any{
		"agent_id":     agent["agent_id"],
		"runtime_id":   agent["runtime_id"],
		"runtime_kind": agent["runtime_kind"],
		"visibility":   agent["visibility"],
	}
}

func agentIDFromCreatePayload(payload map[string]any) string {
	agent, _ := payload["agent"].(map[string]any)
	return stringValue(agent["agent_id"])
}
