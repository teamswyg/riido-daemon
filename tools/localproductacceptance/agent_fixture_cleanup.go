package main

import "net/http"

func cleanupAgentFixtures(client apiClient, base string, fixture agentFixture) []scenario {
	out := make([]scenario, 0, len(fixture.CreatedIDs))
	for index, agentID := range fixture.CreatedIDs {
		id := "local.saas.agent_fixture.cleanup." + intString(index+1)
		out = append(out, apiQuery(client, id, http.MethodDelete,
			base+"/agents/"+agentID, nil, summarizeAgentDelete))
	}
	return out
}

func summarizeAgentDelete(payload map[string]any) map[string]any {
	return map[string]any{
		"agent_id":                    payload["agent_id"],
		"queued_tasks_unassigned":     payload["queued_tasks_unassigned"],
		"running_tasks_force_stopped": payload["running_tasks_force_stopped"],
	}
}
