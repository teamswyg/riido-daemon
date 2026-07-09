package main

import (
	"strings"
	"testing"
)

func TestTaskFlowTaskIDSelectionOrder(t *testing.T) {
	cfg := taskPlanTestConfig(t, "task-configured", "", "", "")
	if got, source := taskFlowTaskID(cfg, nil); got != "task-configured" || source != "configured" {
		t.Fatalf("configured task got %q source %q", got, source)
	}

	cfg = taskPlanTestConfig(t, "", "", "", "")
	discovery := map[string]any{"assigned_agent_profiles": map[string]any{"task-b": nil, "task-a": nil}}
	if got, source := taskFlowTaskID(cfg, discovery); got != "task-a" || source != "assigned-agent-profiles" {
		t.Fatalf("discovered task got %q source %q", got, source)
	}

	if got, source := taskFlowTaskID(cfg, nil); !strings.HasPrefix(got, "local-qa-") || source != "generated" {
		t.Fatalf("generated task got %q source %q", got, source)
	}
}

func TestTaskPlanExplicitAgentsAndCommentFallbacks(t *testing.T) {
	cfg := taskPlanTestConfig(t, "", "agent-explicit-a", "agent-explicit-b", "  custom body  ")
	candidates := taskFlowAgentCandidates(cfg, map[string]any{"agents": []any{
		map[string]any{"agent_id": "agent-payload-a"},
		map[string]any{"agent_id": "agent-payload-b"},
	}}, agentFixture{})
	if candidates[0].AgentID != "agent-explicit-a" || candidates[1].AgentID != "agent-explicit-b" {
		t.Fatalf("candidates=%+v", candidates)
	}
	if got := taskCommentBody(cfg, "task-a"); got != "custom body" {
		t.Fatalf("comment=%q", got)
	}

	cfg = taskPlanTestConfig(t, "", "", "", "")
	if got := taskCommentBody(cfg, "task-a"); got != "local QA thread message for task-a" {
		t.Fatalf("default comment=%q", got)
	}
}

func TestTaskMutationPlanRequiresTwoCandidates(t *testing.T) {
	cfg := taskPlanTestConfig(t, "", "", "", "")
	if plan, ok := taskMutationPlanFor(cfg, map[string]any{"agents": []any{
		map[string]any{"agent_id": "agent-a"},
	}}, "task-a", "configured", agentFixture{}); ok || plan.TaskID != "" {
		t.Fatalf("plan=%+v ok=%v", plan, ok)
	}
}
