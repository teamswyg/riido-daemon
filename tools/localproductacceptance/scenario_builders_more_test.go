package main

import "testing"

func TestContractAPIScenariosSkipsWhenWorkspaceMissing(t *testing.T) {
	empty, token, host := "", "token", "https://example.invalid"
	cfg := config{workspaceID: &empty, apiToken: &token, agentHost: &host}
	scenarios := contractAPIScenarios(cfg)
	if len(scenarios) == 0 {
		t.Fatal("expected skipped contract scenarios")
	}
	for _, sc := range scenarios {
		if sc.Status != statusSkipped || sc.Repair == nil {
			t.Fatalf("scenario should be skipped with repair: %#v", sc)
		}
	}
}

func TestTaskFixtureHelpersPreserveFlowEvidence(t *testing.T) {
	passed := scenario{ID: "team", Status: statusPassed}
	failed := scenario{ID: "create", Status: statusFailed}
	fixture := taskFixture{Team: passed, Create: failed}
	scenarios := taskFixtureScenarios(fixture)
	if len(scenarios) != 2 || !taskFixtureBlocked(fixture) {
		t.Fatalf("fixture scenarios=%#v blocked=%v", scenarios, taskFixtureBlocked(fixture))
	}
	workspace, token, host, enabled := "ws", "tok", "https://example.invalid", false
	cfg := config{workspaceID: &workspace, riidoAPIHost: &host, apiToken: &token, taskFixture: &enabled}
	out := finishTaskFlow(cfg, newAPIClient(host, token), taskFixture{}, agentFixture{}, nil, []scenario{{ID: "tail"}})
	if len(out) != 1 || out[0].ID != "tail" {
		t.Fatalf("unexpected finished flow: %#v", out)
	}
	_ = workspace
}

func TestAssignmentBlockedScenariosPreferRuntimeBindingRepair(t *testing.T) {
	run := taskAssignmentRun{Scenarios: []scenario{{
		ID:     "assignment",
		Status: statusFailed,
		Repair: &repair{Class: "ai_agent_runtime_binding_required"},
	}}}
	scenarios := assignmentBlockedScenarios(run)
	if len(scenarios) != 4 {
		t.Fatalf("blocked scenarios=%#v", scenarios)
	}
	for _, sc := range scenarios {
		if sc.Repair == nil || sc.Repair.Class != "ai_agent_runtime_binding_required" {
			t.Fatalf("expected runtime binding repair: %#v", sc)
		}
	}
}

func TestAssignmentCandidateAnnotationAndIDs(t *testing.T) {
	run := taskAssignmentRun{}
	agent := taskAgentCandidate{AgentID: "agent-a", RuntimeKind: "codex", RuntimeID: "runtime-a"}
	sc := scenario{Status: statusPassed}
	annotateAssignmentCandidate(&sc, agent)
	assignSuccessfulCandidate(&run, agent, sc)
	if candidateAssignmentID(2) != "contract.task.assignment.create.candidate.3" {
		t.Fatal("candidate id drift")
	}
	if run.First.ID != "contract.task.assignment.create.first" ||
		run.First.Observed["candidate_runtime_id"] != "runtime-a" {
		t.Fatalf("run=%#v", run)
	}
	assignSuccessfulCandidate(&run, taskAgentCandidate{AgentID: "agent-b"}, scenario{Status: statusPassed})
	if !run.OK || run.Second.ID != "contract.task.assignment.create.second" {
		t.Fatalf("second assignment not captured: %#v", run)
	}
}
