package main

import (
	"slices"
	"testing"
)

func TestAppendProductTaskArgsEnablesMutations(t *testing.T) {
	taskID, first, second, comment := "task-a", "agent-a", "agent-b", "hello"
	mutations := true
	prepareDaemon := true
	args := appendProductTaskArgs(nil, config{
		productTaskID:        &taskID,
		productAgentID1:      &first,
		productAgentID2:      &second,
		productCommentBody:   &comment,
		productMutations:     &mutations,
		productPrepareDaemon: &prepareDaemon,
	})
	for _, want := range []string{"-task-id", "task-a", "-first-agent-id", "agent-a", "-run-task-mutations", "-prepare-saas-daemon"} {
		if !slices.Contains(args, want) {
			t.Fatalf("args missing %q: %v", want, args)
		}
	}
}
