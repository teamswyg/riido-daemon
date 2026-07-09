package main

import (
	"strings"
	"testing"
)

func TestAgentFixtureSummariesAndIdentityContract(t *testing.T) {
	payload := map[string]any{"agent": map[string]any{
		"agent_id": "agent_1", "runtime_id": "runtime_1",
		"runtime_kind": "codex", "visibility": "private",
	}}
	summary := summarizeAgentCreate(payload)
	if summary["agent_id"] != "agent_1" || summary["runtime_kind"] != "codex" {
		t.Fatalf("unexpected create summary: %#v", summary)
	}
	if agentIDFromCreatePayload(payload) != "agent_1" {
		t.Fatal("agent id should be extracted from create payload")
	}
	deleted := summarizeAgentDelete(map[string]any{
		"agent_id": "agent_1", "queued_tasks_unassigned": 2,
		"running_tasks_force_stopped": 1,
	})
	if deleted["running_tasks_force_stopped"] != 1 {
		t.Fatalf("unexpected delete summary: %#v", deleted)
	}
	name := agentFixtureName(2, "claude")
	if !strings.HasPrefix(name, "Local QA claude 2 ") {
		t.Fatalf("unexpected fixture name: %s", name)
	}
	if agentCreateScenarioID(3) != "local.saas.agent_fixture.create.3" {
		t.Fatal("unexpected create scenario id")
	}
}

func TestAgentIdentityContractRequiresTwoDistinctAgents(t *testing.T) {
	fixture := agentFixture{Candidates: []taskAgentCandidate{
		{AgentID: "agent_1", RuntimeKind: "claude"},
		{AgentID: "agent_2", RuntimeKind: "claude"},
	}}
	pair := [2]preparedRuntime{{RuntimeID: "rt_1"}, {RuntimeID: "rt_2"}}
	sc := agentIdentityContractScenario(fixture, pair)
	if sc.Status != statusPassed || sc.Observed["forbidden_dedupe"] == "" {
		t.Fatalf("expected passed identity contract, got %#v", sc)
	}
	sc = agentIdentityContractScenario(agentFixture{}, pair)
	if sc.Status != statusFailed || sc.FailureSummary == "" {
		t.Fatalf("expected missing fixture failure, got %#v", sc)
	}
	skipped := agentFixtureSkippedScenario()
	if skipped.Status != statusSkipped || skipped.Repair == nil {
		t.Fatalf("unexpected skipped scenario: %#v", skipped)
	}
}
