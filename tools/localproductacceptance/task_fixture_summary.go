package main

func summarizeTaskFixtureTeams(payload map[string]any) map[string]any {
	return map[string]any{
		"teams_count": arrayLen(payload["teams"]) + arrayLen(payload["data"]) + arrayLen(payload["items"]),
	}
}

func summarizeTaskFixtureCreate(payload map[string]any) map[string]any {
	return map[string]any{
		"task_id_present": firstString(payload, "componentId", "component_id", "id") != "",
	}
}

func summarizeTaskFixtureCleanup(map[string]any) map[string]any {
	return map[string]any{"cleanup_requested": true}
}
