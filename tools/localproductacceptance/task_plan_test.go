package main

import "testing"

func TestTaskMutationPlanUsesGeneratedTaskAndAssignableAgents(t *testing.T) {
	cfg := taskPlanTestConfig(t, "", "", "", "")
	plan, ok := taskMutationPlanFor(cfg, map[string]any{"agents": []any{
		map[string]any{"agent_id": "agent-a", "runtime_kind": "codex"},
		map[string]any{"agent_id": "agent-b", "runtime_kind": "codex"},
	}}, "task-a", "test", agentFixture{})
	if !ok {
		t.Fatal("plan not selected")
	}
	if plan.TaskIDSource != "test" || plan.TaskID != "task-a" {
		t.Fatalf("plan task = %+v", plan)
	}
	if plan.Pair.First.AgentID != "agent-a" || plan.Pair.Second.AgentID != "agent-b" {
		t.Fatalf("pair=%+v", plan.Pair)
	}
	if plan.CommentBody == "" {
		t.Fatalf("comment body missing: %+v", plan)
	}
}

func TestTaskMutationPlanPrefersPreparedAgentFixture(t *testing.T) {
	cfg := taskPlanTestConfig(t, "", "", "", "")
	fixture := agentFixture{Candidates: []taskAgentCandidate{
		{AgentID: "agent-prepared-a", RuntimeKind: "codex"},
		{AgentID: "agent-prepared-b", RuntimeKind: "codex"},
	}}
	plan, ok := taskMutationPlanFor(cfg, nil, "task-a", "test", fixture)
	if !ok || plan.Pair.First.AgentID != "agent-prepared-a" {
		t.Fatalf("plan=%+v ok=%v", plan, ok)
	}
}

func taskPlanTestConfig(t *testing.T, taskID, first, second, comment string) config {
	t.Helper()
	return config{
		taskID:        &taskID,
		firstAgentID:  &first,
		secondAgentID: &second,
		commentBody:   &comment,
	}
}
