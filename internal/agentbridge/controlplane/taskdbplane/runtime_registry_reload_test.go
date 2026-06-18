package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestNewLoadsPersistedRuntimeRegistryForClaimGating(t *testing.T) {
	path := writeTaskDB(t, taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
			{ID: "claude", Available: true},
		},
		Tasks: []taskdb.TaskRecord{
			{
				ID:                  "codex-task",
				ProjectID:           "project-1",
				State:               task.StateQueued,
				Title:               "codex task",
				RecommendedProvider: "codex",
				UpdatedAt:           "2026-05-25T00:00:00Z",
			},
			{
				ID:                  "claude-task",
				ProjectID:           "project-1",
				State:               task.StateQueued,
				Title:               "claude task",
				RecommendedProvider: "claude",
				UpdatedAt:           "2026-05-25T00:00:01Z",
			},
		},
	})
	plane := newTestPlane(t, path)
	err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		RuntimeID: "runtime-claude",
		Provider:  "multi",
		Capabilities: map[string]bool{
			"provider.codex.available":  false,
			"provider.claude.available": true,
		},
	})
	if err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}

	reloaded, err := New(path)
	if err != nil {
		t.Fatalf("New reload: %v", err)
	}
	reloaded.now = plane.now
	req, err := reloaded.ClaimTask(context.Background(), "runtime-claude")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req == nil || req.ID != "claude-task" {
		t.Fatalf("reloaded runtime registry should gate claim, got %+v", req)
	}
}
