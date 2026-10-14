package project

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestSyncTaskDBFromStateUpdatesOrchestrationMetadata(t *testing.T) {
	state := sampleTaskDBState(t)
	first := SyncTaskDBFromState(taskdb.EmptyTaskDB(), state, taskDBSyncTime(10))
	state = taskDBStateWithClaudeRecommendation(state)

	second := SyncTaskDBFromState(first, state, taskDBSyncTime(12))

	if second.RecommendedProvider != "claude" {
		t.Fatalf("sync should update DB recommended provider: %s", second.RecommendedProvider)
	}
	if len(second.ProviderCandidates) != 1 || second.ProviderCandidates[0].ID != "claude" {
		t.Fatalf("sync should update provider candidates: %#v", second.ProviderCandidates)
	}

	record := findTaskRecord(t, second, "task:mws.goal")
	if record.RecommendedProvider != "claude" {
		t.Fatalf("sync should update task recommended provider: %s", record.RecommendedProvider)
	}
	if len(second.Transitions) != len(first.Transitions) {
		t.Fatalf(
			"metadata sync should not append transitions: first=%d second=%d",
			len(first.Transitions),
			len(second.Transitions),
		)
	}
}

func taskDBStateWithClaudeRecommendation(state StateFile) StateFile {
	state.RecommendedProvider = "claude"
	state.RecommendedDecisionLLM = "codex"
	state.ProviderCandidates = []ProviderCandidate{
		{
			ID:               "claude",
			SourceWorkflow:   "provider-selection",
			Available:        true,
			ApprovalRequired: true,
		},
	}
	for index := range state.Tasks {
		state.Tasks[index].RecommendedProvider = "claude"
		state.Tasks[index].RecommendedDecisionLLM = "codex"
	}

	return state
}
