package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestClaimTaskSkipsUnavailableProviderForRuntime(t *testing.T) {
	path := writeTaskDB(t, providerClaimFixture())
	plane := newTestPlane(t, path)
	if err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		RuntimeID: "runtime-claude",
		Provider:  "multi",
		Capabilities: map[string]bool{
			"provider.codex.available":  false,
			"provider.claude.available": true,
		},
	}); err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}

	req, err := plane.ClaimTask(context.Background(), "runtime-claude")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req == nil || req.ID != "claude-task" {
		t.Fatalf("runtime should skip unavailable provider and claim claude task, got %+v", req)
	}
	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "codex-task").State; got != task.StateQueued {
		t.Fatalf("codex task should remain queued, got %s", got)
	}
	if got := mustFindTask(t, db, "claude-task").State; got != task.StateClaimed {
		t.Fatalf("claude task should be claimed, got %s", got)
	}
}

func providerClaimFixture() taskdb.TaskDB {
	return taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
			{ID: "claude", Available: true},
		},
		Tasks: []taskdb.TaskRecord{
			queuedProviderTask("codex-task", "codex", "2026-05-25T00:00:00Z"),
			queuedProviderTask("claude-task", "claude", "2026-05-25T00:00:01Z"),
		},
	}
}

func queuedProviderTask(id, provider, updatedAt string) taskdb.TaskRecord {
	return taskdb.TaskRecord{
		ID:                  id,
		ProjectID:           "project-1",
		State:               task.StateQueued,
		Title:               provider + " task",
		RecommendedProvider: provider,
		UpdatedAt:           updatedAt,
	}
}
