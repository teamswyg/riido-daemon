package project

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestSyncTaskDBFromStateCopiesTaskRecommendations(t *testing.T) {
	state := sampleTaskDBState(t)

	db := SyncTaskDBFromState(taskdb.EmptyTaskDB(), state, taskDBSyncTime(10))

	record := findTaskRecord(t, db, "task:mws.goal")
	if record.RecommendedProvider != "codex" {
		t.Fatalf("unexpected task recommended provider: %s", record.RecommendedProvider)
	}
	if record.RecommendedDecisionLLM != "codex" {
		t.Fatalf("unexpected task decision LLM: %s", record.RecommendedDecisionLLM)
	}
	if !record.RequiresHumanApproval {
		t.Fatalf("task should require human approval")
	}
}
