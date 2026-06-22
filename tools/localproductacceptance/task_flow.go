package main

import "net/http"

func taskFlowScenarios(client apiClient, cfg config) []scenario {
	base := workspaceBase(*cfg.workspaceID)
	if *cfg.taskID == "" {
		return taskSkipped(true, "Set RIIDO_E2E_TASK_ID to verify a real task flow.")
	}
	out := []scenario{
		apiQuery(client, "contract.task.assignable_agents", http.MethodGet,
			taskEndpoint(base, *cfg.taskID, "/assignable-agents"), nil, summarizeAssignableAgents),
	}
	if !*cfg.runMutations {
		return append(out, taskSkipped(false, "Pass -run-task-mutations with two agent ids.")...)
	}
	return append(out, taskMutationScenarios(client, cfg, base)...)
}

func taskMutationScenarios(client apiClient, cfg config, base string) []scenario {
	if *cfg.firstAgentID == "" || *cfg.secondAgentID == "" {
		return taskSkipped(false, "Set RIIDO_E2E_AGENT_ID_1 and RIIDO_E2E_AGENT_ID_2.")
	}
	first := createAssignment(client, "contract.task.assignment.create.first", base, *cfg.taskID, *cfg.firstAgentID)
	second := createAssignment(client, "contract.task.assignment.create.second", base, *cfg.taskID, *cfg.secondAgentID)
	out := []scenario{first, second, distinctAssignmentScenario(first, second)}
	out = append(out, apiQuery(client, "contract.task.thread_subscription", http.MethodGet,
		taskEndpoint(base, *cfg.taskID, "/thread-stream-subscription"), nil, summarizeSubscription))
	out = append(out, maybeThreadMessage(client, cfg, base, first))
	return out
}

func createAssignment(client apiClient, id, base, taskID, agentID string) scenario {
	body := map[string]any{"agent_id": agentID}
	return apiQuery(client, id, http.MethodPost,
		taskEndpoint(base, taskID, "/agent-assignments"), body, summarizeTaskAction)
}

func maybeThreadMessage(client apiClient, cfg config, base string, assigned scenario) scenario {
	threadID, _ := assigned.Observed["thread_id"].(string)
	if *cfg.commentBody == "" || threadID == "" {
		return failTaskScenario("contract.task.thread_message", "Set -comment-body after assignment creates a thread.")
	}
	path := taskEndpoint(base, *cfg.taskID, "/threads/"+threadID+"/messages")
	return apiQuery(client, "contract.task.thread_message", http.MethodPost,
		path, map[string]any{"body": *cfg.commentBody}, summarizeTaskAction)
}

func taskSkipped(includeAssignable bool, summary string) []scenario {
	out := []scenario{
		failTaskScenario("contract.task.multi_assignment", summary),
		failTaskScenario("contract.task.thread_subscription", summary),
		failTaskScenario("contract.task.thread_message", summary),
	}
	if includeAssignable {
		out = append([]scenario{failTaskScenario("contract.task.assignable_agents", summary)}, out...)
	}
	return out
}
