package project

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestSyncTaskDBFromStateDoesNotDuplicateInitialTransitions(t *testing.T) {
	state := sampleTaskDBState(t)
	first := SyncTaskDBFromState(taskdb.EmptyTaskDB(), state, taskDBSyncTime(10))
	second := SyncTaskDBFromState(first, state, taskDBSyncTime(11))

	if len(second.Transitions) != len(first.Transitions) {
		t.Fatalf(
			"sync should not duplicate initial transitions: first=%d second=%d",
			len(first.Transitions),
			len(second.Transitions),
		)
	}

	record := findTaskRecord(t, second, "task:mws.goal")
	if record.TransitionCount != 1 {
		t.Fatalf("unexpected transition count after second sync: %d", record.TransitionCount)
	}
}
