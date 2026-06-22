package main

import "net/http"

func taskFlowScenarios(client apiClient, cfg config, discovery map[string]any) []scenario {
	base := workspaceBase(*cfg.workspaceID)
	taskID, source := taskFlowTaskID(cfg, discovery)
	assignable, payload := apiQueryPayload(client, "contract.task.assignable_agents", http.MethodGet,
		taskEndpoint(base, taskID, "/assignable-agents"), nil, summarizeAssignableAgents)
	assignable.Observed["task_id"] = taskID
	assignable.Observed["task_id_source"] = source
	out := []scenario{assignable}
	if shouldSkipGeneratedTaskFlow(assignable, source) {
		summary := "Set RIIDO_E2E_TASK_ID to a real accessible task; generated task was rejected."
		out[0] = skippedTaskScenario(assignable, summary)
		return append(out, taskSkipped(false, summary)...)
	}
	if !*cfg.runMutations {
		return append(out, taskSkipped(false, "Pass -run-task-mutations with two agent ids.")...)
	}
	plan, ok := taskMutationPlanFor(cfg, payload, taskID, source)
	if !ok {
		return append(out, taskSkipped(false, "Need at least two assignable AI agents.")...)
	}
	return append(out, taskMutationScenarios(client, base, plan)...)
}

func shouldSkipGeneratedTaskFlow(assignable scenario, source string) bool {
	return source == "generated" && assignable.Status == statusFailed
}

func taskMutationScenarios(client apiClient, base string, plan taskMutationPlan) []scenario {
	first := createAssignment(client, "contract.task.assignment.create.first", base, plan.TaskID, plan.Pair.First.AgentID)
	second := createAssignment(client, "contract.task.assignment.create.second", base, plan.TaskID, plan.Pair.Second.AgentID)
	out := []scenario{first, second, distinctAssignmentScenario(plan, first, second)}
	out = append(out, apiQuery(client, "contract.task.thread_subscription", http.MethodGet,
		taskEndpoint(base, plan.TaskID, "/thread-stream-subscription"), nil, summarizeSubscription))
	out = append(out, sseReplayScenario(client, base, first, second))
	out = append(out, threadMessageScenario(client, base, plan, first))
	out = append(out, cleanupTaskAssignments(client, base, plan)...)
	return out
}

func createAssignment(client apiClient, id, base, taskID, agentID string) scenario {
	body := map[string]any{"agent_id": agentID}
	return apiQuery(client, id, http.MethodPost,
		taskEndpoint(base, taskID, "/agent-assignments"), body, summarizeTaskAction)
}

func taskSkipped(includeAssignable bool, summary string) []scenario {
	out := []scenario{
		failTaskScenario("contract.task.multi_assignment", summary),
		failTaskScenario("contract.task.thread_subscription", summary),
		failTaskScenario("contract.task.sse_replay", summary),
		failTaskScenario("contract.task.thread_message", summary),
	}
	if includeAssignable {
		out = append([]scenario{failTaskScenario("contract.task.assignable_agents", summary)}, out...)
	}
	return out
}
