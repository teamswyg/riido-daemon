package taskdbplane

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestClaimTaskSkipsActiveLeaseOwnedByAnotherRuntime(t *testing.T) {
	path := writeTaskDB(t, taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                  "codex-task",
			State:               task.StateQueued,
			Title:               "codex task",
			RecommendedProvider: "codex",
		}},
	})
	plane := newTestPlane(t, path)
	registerRuntimeForProvider(t, plane, "runtime-1", "codex", 2, 0)
	now := time.Date(2026, 5, 25, 2, 0, 0, 0, time.UTC)
	if err := saveRuntimeLeaseRegistry(plane.leasePath, path, RuntimeLeaseRegistry{Leases: []RuntimeLeaseRecord{{
		LeaseID:               "runtime-lease:codex-task:7",
		TaskID:                "codex-task",
		RuntimeID:             "runtime-other",
		CapabilityFingerprint: "runtime-1-fp",
		ClaimedAt:             now,
		LeaseUntil:            now.Add(time.Hour),
		FencingToken:          7,
	}}}, now); err != nil {
		t.Fatalf("saveRuntimeLeaseRegistry: %v", err)
	}

	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req != nil {
		t.Fatalf("task with active foreign lease should not be claimed: %+v", req)
	}
	if got := mustFindTask(t, loadTaskDB(t, path), "codex-task").State; got != task.StateQueued {
		t.Fatalf("task should remain queued, got %s", got)
	}
	registry := readRuntimeLeaseRegistry(t, plane.leasePath)
	if len(registry.Leases) != 1 || registry.Leases[0].FencingToken != 7 {
		t.Fatalf("foreign lease should remain unchanged: %+v", registry.Leases)
	}
}

func TestClaimTaskReclaimsExpiredLeaseWithIncrementedFencingToken(t *testing.T) {
	path := writeTaskDB(t, taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                  "codex-task",
			State:               task.StateQueued,
			Title:               "codex task",
			RecommendedProvider: "codex",
		}},
	})
	plane := newTestPlane(t, path)
	registerRuntimeForProvider(t, plane, "runtime-1", "codex", 2, 0)
	now := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	if err := saveRuntimeLeaseRegistry(plane.leasePath, path, RuntimeLeaseRegistry{Leases: []RuntimeLeaseRecord{{
		LeaseID:               "runtime-lease:codex-task:7",
		TaskID:                "codex-task",
		RuntimeID:             "runtime-other",
		CapabilityFingerprint: "runtime-other-fp",
		ClaimedAt:             now.Add(-2 * time.Hour),
		LeaseUntil:            now.Add(-time.Hour),
		FencingToken:          7,
	}}}, now); err != nil {
		t.Fatalf("saveRuntimeLeaseRegistry: %v", err)
	}

	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req == nil || req.Metadata[controlplane.MetadataRuntimeFencingToken] != "8" {
		t.Fatalf("expired lease should be reclaimed with token 8, got %+v", req)
	}
	registry := readRuntimeLeaseRegistry(t, plane.leasePath)
	if len(registry.Leases) != 1 || registry.Leases[0].RuntimeID != "runtime-1" || registry.Leases[0].FencingToken != 8 {
		t.Fatalf("reclaimed lease mismatch: %+v", registry.Leases)
	}
}

func TestClaimTaskRequeuesExpiredRunningLeaseForHandoff(t *testing.T) {
	path := writeTaskDB(t, taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                  "codex-task",
			State:               task.StateRunning,
			Title:               "codex task",
			RecommendedProvider: "codex",
		}},
	})
	plane := newTestPlane(t, path)
	registerRuntimeForProvider(t, plane, "runtime-new", "codex", 2, 0)
	now := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	writeRuntimeLease(t, plane, RuntimeLeaseRecord{
		LeaseID:               "runtime-lease:codex-task:3",
		TaskID:                "codex-task",
		RuntimeID:             "runtime-old",
		CapabilityFingerprint: "runtime-old-fp",
		ClaimedAt:             now.Add(-2 * time.Hour),
		LeaseUntil:            now.Add(-time.Hour),
		FencingToken:          3,
	})

	req, err := plane.ClaimTask(context.Background(), "runtime-new")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req == nil || req.ID != "codex-task" || req.Metadata[controlplane.MetadataRuntimeFencingToken] != "4" {
		t.Fatalf("expired running lease should be handed off and claimed with token 4, got %+v", req)
	}
	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "codex-task").State; got != task.StateClaimed {
		t.Fatalf("task should be claimed by new runtime after handoff, got %s", got)
	}
	assertTransition(t, db, ir.EventBlockerRaised)
	assertTransition(t, db, ir.EventBlockerResolvedRequeue)
	assertTransition(t, db, ir.EventTaskClaimed)
	registry := readRuntimeLeaseRegistry(t, plane.leasePath)
	if len(registry.Leases) != 1 || registry.Leases[0].RuntimeID != "runtime-new" || registry.Leases[0].FencingToken != 4 {
		t.Fatalf("new lease mismatch after handoff: %+v", registry.Leases)
	}
}

func TestClaimTaskFailsExpiredClaimedLease(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	now := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	writeRuntimeLease(t, plane, RuntimeLeaseRecord{
		LeaseID:               "runtime-lease:task-1:5",
		TaskID:                "task-1",
		RuntimeID:             "runtime-old",
		CapabilityFingerprint: "runtime-old-fp",
		ClaimedAt:             now.Add(-2 * time.Hour),
		LeaseUntil:            now.Add(-time.Hour),
		FencingToken:          5,
	})

	req, err := plane.ClaimTask(context.Background(), "runtime-new")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req != nil {
		t.Fatalf("expired claimed lease should fail task rather than hand off: %+v", req)
	}
	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-1").State; got != task.StateFailed {
		t.Fatalf("task should fail after expired Claimed lease, got %s", got)
	}
	assertTransition(t, db, ir.EventTaskFailed)
	registry := readRuntimeLeaseRegistry(t, plane.leasePath)
	if len(registry.Leases) != 1 || registry.Leases[0].ReleasedAt == nil {
		t.Fatalf("expired claimed lease should be released: %+v", registry.Leases)
	}
}
