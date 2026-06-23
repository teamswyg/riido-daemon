package main

import "net/http"

func taskFlowScenarios(client apiClient, cfg config, discovery map[string]any, agents agentFixture) []scenario {
	base := workspaceBase(*cfg.workspaceID)
	taskID, source := taskFlowTaskID(cfg, discovery)
	fixture := maybeCreateTaskFixture(cfg, source)
	out := taskFixtureScenarios(fixture)
	if taskFixtureBlocked(fixture) {
		fallback := existingTaskFallback(newAPIClient(*cfg.riidoAPIHost, *cfg.apiToken), fixture.TeamID)
		if fallback.Scenario.ID != "" {
			out = append(out, fallback.Scenario)
		}
		if fallback.TaskID != "" {
			out = markFixtureFallback(out, fallback)
			taskID, source = fallback.TaskID, "readable-task-fallback"
		} else {
			summary := "Create a staging task fixture or set RIIDO_E2E_TASK_ID."
			tail := taskSkipped(false, summary)
			return finishTaskFlow(cfg, client, fixture, agents, out, tail)
		}
	}
	if fixture.Created() {
		taskID, source = fixture.TaskID, "created-fixture"
	}
	assignable, payload := apiQueryPayload(client, "contract.task.assignable_agents", http.MethodGet,
		taskEndpoint(base, taskID, "/assignable-agents"), nil, summarizeAssignableAgents)
	assignable.Observed["task_id"] = taskID
	assignable.Observed["task_id_source"] = source
	out = append(out, assignable)
	if shouldSkipGeneratedTaskFlow(assignable, source) {
		summary := "Set RIIDO_E2E_TASK_ID to a real accessible task; generated task was rejected."
		out[0] = skippedTaskScenario(assignable, summary)
		tail := taskSkipped(false, summary)
		return finishTaskFlow(cfg, client, fixture, agents, out, tail)
	}
	if !*cfg.runMutations {
		return append(out, taskSkipped(false, "Pass -run-task-mutations with two agent ids.")...)
	}
	plan, ok := taskMutationPlanFor(cfg, payload, taskID, source, agents)
	if !ok {
		tail := taskSkipped(false, "Need at least two assignable AI agents.")
		return finishTaskFlow(cfg, client, fixture, agents, out, tail)
	}
	tail := taskMutationScenarios(client, base, plan)
	return finishTaskFlow(cfg, client, fixture, agents, out, tail)
}

func shouldSkipGeneratedTaskFlow(assignable scenario, source string) bool {
	return source == "generated" && assignable.Status == statusFailed
}

func taskMutationScenarios(client apiClient, base string, plan taskMutationPlan) []scenario {
	run := createAssignmentRun(client, base, plan)
	out := append([]scenario{}, run.Scenarios...)
	if !run.OK {
		out = append(out, assignmentBlockedScenarios(run)...)
		return append(out, cleanupPartialTaskAssignments(client, base, plan.TaskID, run)...)
	}
	plan.Pair = run.Pair
	out = append(out, distinctAssignmentScenario(plan, run.First, run.Second))
	out = append(out, apiQuery(client, "contract.task.thread_subscription", http.MethodGet,
		taskEndpoint(base, plan.TaskID, "/thread-stream-subscription"), nil, summarizeSubscription))
	out = append(out, sseReplayScenario(client, base, run.First, run.Second))
	out = append(out, threadMessageScenario(client, base, plan, run.First))
	out = append(out, cleanupTaskAssignments(client, base, plan)...)
	return out
}

func createAssignment(client apiClient, id, base, taskID, agentID string) scenario {
	body := map[string]any{"agent_id": agentID}
	return apiQuery(client, id, http.MethodPost,
		taskEndpoint(base, taskID, "/agent-assignments"), body, summarizeTaskAction)
}
