package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestClaimTaskSkipsUnavailableProviderForRuntime(t *testing.T) {
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

func TestClaimTaskSelectsRuntimePoolCandidate(t *testing.T) {
	path := writeTaskDB(t, taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                  "codex-task",
			ProjectID:           "project-1",
			State:               task.StateQueued,
			Title:               "codex task",
			RecommendedProvider: "codex",
		}},
	})
	plane := newTestPlane(t, path)
	registerRuntimeForProvider(t, plane, "runtime-full", "codex", 1, 1)
	registerRuntimeForProvider(t, plane, "runtime-free", "codex", 2, 0)

	first, err := plane.ClaimTask(context.Background(), "runtime-full")
	if err != nil {
		t.Fatalf("ClaimTask runtime-full returned error: %v", err)
	}
	if first != nil {
		t.Fatalf("slot-exhausted runtime should not claim selected task: %+v", first)
	}
	if got := mustFindTask(t, loadTaskDB(t, path), "codex-task").State; got != task.StateQueued {
		t.Fatalf("task should remain queued after non-selected runtime claim, got %s", got)
	}

	second, err := plane.ClaimTask(context.Background(), "runtime-free")
	if err != nil {
		t.Fatalf("ClaimTask runtime-free returned error: %v", err)
	}
	if second == nil || second.ID != "codex-task" {
		t.Fatalf("selected runtime should claim task, got %+v", second)
	}
	if got := mustFindTask(t, loadTaskDB(t, path), "codex-task").State; got != task.StateClaimed {
		t.Fatalf("task should be claimed by selected runtime, got %s", got)
	}
}

func TestClaimTaskPersistsRuntimeLeaseMetadata(t *testing.T) {
	path := writeTaskDB(t, taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                  "codex-task",
			ProjectID:           "project-1",
			State:               task.StateQueued,
			Title:               "codex task",
			RecommendedProvider: "codex",
		}},
	})
	plane := newTestPlane(t, path)
	registerRuntimeForProvider(t, plane, "runtime-1", "codex", 2, 0)

	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req == nil {
		t.Fatal("ClaimTask returned nil request")
		return
	}
	if req.Metadata[controlplane.MetadataRuntimeLeaseID] != "runtime-lease:codex-task:1" {
		t.Fatalf("lease metadata missing: %+v", req.Metadata)
	}
	if req.Metadata[controlplane.MetadataRuntimeFencingToken] != "1" {
		t.Fatalf("fencing token metadata missing: %+v", req.Metadata)
	}
	if req.Metadata[controlplane.MetadataRuntimeCapabilityFingerprint] != "runtime-1-fp" {
		t.Fatalf("capability fingerprint metadata missing: %+v", req.Metadata)
	}
	registry := readRuntimeLeaseRegistry(t, plane.leasePath)
	if registry.SchemaVersion != RuntimeLeaseRegistrySchemaVersion || registry.TaskDBPath != path {
		t.Fatalf("lease registry identity mismatch: %+v", registry)
	}
	if len(registry.Leases) != 1 {
		t.Fatalf("lease count = %d, want 1: %+v", len(registry.Leases), registry.Leases)
	}
	lease := registry.Leases[0]
	if lease.TaskID != "codex-task" || lease.RuntimeID != "runtime-1" || lease.FencingToken != 1 {
		t.Fatalf("lease mismatch: %+v", lease)
	}
	if lease.CapabilityFingerprint != "runtime-1-fp" || lease.ReleasedAt != nil {
		t.Fatalf("lease fingerprint/release mismatch: %+v", lease)
	}
	if !lease.LeaseUntil.After(lease.ClaimedAt) {
		t.Fatalf("lease deadline should be after claim time: %+v", lease)
	}
}
