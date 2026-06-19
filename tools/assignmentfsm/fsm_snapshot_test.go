package main

import (
	"slices"
	"testing"
)

func TestAssignmentFSMSnapshotComesFromContracts(t *testing.T) {
	fsm := buildFSMSnapshot()
	if fsm.Name != "assignment" || fsm.TypeUnion != "AssignmentPollingFSM" {
		t.Fatalf("unexpected fsm identity: %+v", fsm)
	}
	if !slices.Contains(fsm.States, "ready") || slices.Contains(fsm.States, "preparing_workspace") {
		t.Fatalf("fsm states must reflect contracts, got %+v", fsm.States)
	}
	if !slices.Contains(fsm.TerminalStates, "completed") || slices.Contains(fsm.TerminalStates, "blocked") {
		t.Fatalf("terminal states must reflect contracts, got %+v", fsm.TerminalStates)
	}
}
