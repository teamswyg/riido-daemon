package main

import (
	"strings"
	"testing"
)

func TestAppendScheduleProductArgsIncludesTaskFlags(t *testing.T) {
	cfg := uploadTestConfig("", "", ".riido-local/coverage.json", "", "", "")
	runProduct, startClient, mutations, fixture := false, true, false, false
	workspace, team := "workspace-a", "team-a"
	task, first, second, comment := "task-a", "agent-a", "agent-b", "hello"
	cfg.runProduct = &runProduct
	cfg.productStartClient = &startClient
	cfg.productMutations = &mutations
	cfg.productTaskFixture = &fixture
	cfg.productWorkspace = &workspace
	cfg.productTeamID = &team
	cfg.productTaskID = &task
	cfg.productAgentID1 = &first
	cfg.productAgentID2 = &second
	cfg.productCommentBody = &comment
	got := appendScheduleProductArgs([]string{}, cfg)
	joined := joinArgs(got)
	for _, want := range []string{
		"-run-product=false",
		"-product-start-client",
		"-product-workspace-id workspace-a",
		"-product-team-id team-a",
		"-product-task-id task-a",
		"-product-agent-id-1 agent-a",
		"-product-agent-id-2 agent-b",
		"-product-comment-body hello",
		"-product-task-mutations=false",
		"-product-create-task-fixture=false",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("args missing %q: %v", want, got)
		}
	}
}

func joinArgs(args []string) string {
	return strings.Join(args, " ")
}
