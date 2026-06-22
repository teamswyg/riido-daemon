package main

import "testing"

func TestChooseTaskAgentPairPrefersSameRuntimeKind(t *testing.T) {
	pair, ok := chooseTaskAgentPair([]taskAgentCandidate{
		{AgentID: "agent-a", RuntimeKind: "codex"},
		{AgentID: "agent-b", RuntimeKind: "claude_code"},
		{AgentID: "agent-c", RuntimeKind: "codex"},
	})
	if !ok {
		t.Fatal("pair not selected")
	}
	if pair.First.AgentID != "agent-a" || pair.Second.AgentID != "agent-c" {
		t.Fatalf("pair=%+v", pair)
	}
}

func TestChooseTaskAgentPairFallsBackToFirstTwo(t *testing.T) {
	pair, ok := chooseTaskAgentPair([]taskAgentCandidate{
		{AgentID: "agent-a", RuntimeKind: "codex"},
		{AgentID: "agent-b", RuntimeKind: "openclaw"},
	})
	if !ok || pair.First.AgentID != "agent-a" || pair.Second.AgentID != "agent-b" {
		t.Fatalf("pair=%+v ok=%v", pair, ok)
	}
}

func TestPrioritizeTaskAgentCandidatesKeepsSameRuntimeKindFirst(t *testing.T) {
	agents := prioritizeTaskAgentCandidates([]taskAgentCandidate{
		{AgentID: "agent-a", RuntimeKind: "codex"},
		{AgentID: "agent-b", RuntimeKind: "openclaw"},
		{AgentID: "agent-c", RuntimeKind: "codex"},
	})
	if agents[0].AgentID != "agent-a" || agents[1].AgentID != "agent-c" {
		t.Fatalf("agents=%+v", agents)
	}
	if agents[2].AgentID != "agent-b" {
		t.Fatalf("fallback candidate lost: %+v", agents)
	}
}
