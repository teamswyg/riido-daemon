package taskdbplane

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestRuntimeCapabilityForProviderReadsWorktreeSurface(t *testing.T) {
	prefix := "provider.openclaw."
	capability, ok := runtimeCapabilityForProvider(controlplane.RegisteredRuntime{
		RuntimeRegistration: controlplane.RuntimeRegistration{
			RuntimeID: "runtime-1",
			Capabilities: map[string]bool{
				prefix + "available":                    true,
				prefix + "requires_experimental_opt_in": true,
				prefix + "supports_streaming":           true,
				prefix + "supports_resume":              true,
				prefix + "supports_usage":               true,
				prefix + "supports_worktree":            false,
			},
			CapabilityAttributes: map[string]string{
				prefix + "compatibility_status":   "experimental",
				prefix + "capability_fingerprint": "fp-openclaw",
			},
		},
	}, "openclaw")
	if !ok {
		t.Fatal("expected provider capability")
	}
	if capability.SupportsWorktree {
		t.Fatalf("worktree support must mirror runtime registry, got %+v", capability)
	}
	if !capability.SupportsStreaming || !capability.SupportsResume || !capability.SupportsUsage {
		t.Fatalf("other support flags not preserved: %+v", capability)
	}
}

func TestClaimTaskTransitionsQueuedRowAndBuildsRequest(t *testing.T) {
	path := writeTaskDB(t, taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                   "task-1",
			ProjectID:            "project-1",
			State:                task.StateQueued,
			Title:                "fallback title",
			RecommendedProvider:  "codex",
			HarnessNextDirection: "implement the patch",
			SourceDocumentPath:   "docs/task.md",
			UpdatedAt:            "2026-05-25T00:00:00Z",
		}},
	})
	plane := newTestPlane(t, path)

	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req == nil || req.ID != "task-1" || req.Provider != "codex" || req.Prompt != "implement the patch" {
		t.Fatalf("unexpected request: %+v", req)
	}
	if req.Metadata["workspace_id"] != "project-1" {
		t.Fatalf("workspace metadata missing: %+v", req.Metadata)
	}
	if req.Metadata[metadataTaskDB] != path || req.Metadata[metadataDocument] != "docs/task.md" {
		t.Fatalf("task metadata mismatch: %+v", req.Metadata)
	}

	db := loadTaskDB(t, path)
	record := mustFindTask(t, db, "task-1")
	if record.State != task.StateClaimed {
		t.Fatalf("state = %s, want Claimed", record.State)
	}
	if len(db.CommandReceipts) != 1 || db.CommandReceipts[0].CommandID != commandIDPrefix+"task-1:claim:runtime-1" {
		t.Fatalf("claim receipt mismatch: %+v", db.CommandReceipts)
	}
	second, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("second ClaimTask returned error: %v", err)
	}
	if second != nil {
		t.Fatalf("claimed same task twice: %+v", second)
	}
}

func TestClaimTaskReusesExistingApprovalIDForHumanGatedTask(t *testing.T) {
	now := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	db := taskdb.TaskDB{
		SchemaVersion:          taskdb.TaskDBSchemaVersion,
		DecisionGate:           "human-approval-required",
		RecommendedProvider:    "codex",
		RecommendedDecisionLLM: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                     "task-human",
			ProjectID:              "project-1",
			State:                  task.StateCreated,
			Title:                  "approved task",
			RecommendedProvider:    "codex",
			RecommendedDecisionLLM: "codex",
			RequiresHumanApproval:  true,
		}},
	}
	var err error
	db, _, _, err = taskdb.ApplyGuardedTaskTransition(db, taskdb.TaskTransitionInput{
		TaskID:  "task-human",
		ToState: task.StateQueued,
		Event:   ir.EventTaskQueued,
		Actor:   "human",
		Source:  "test",
		Reason:  "approved for run",
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   "command:test:queue",
			Provider:    "codex",
			DecisionLLM: "codex",
			ApprovalID:  "approval:human:1",
		},
	}, now)
	if err != nil {
		t.Fatalf("queue transition: %v", err)
	}
	path := writeTaskDB(t, db)
	plane := newTestPlane(t, path)

	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req == nil || req.ID != "task-human" {
		t.Fatalf("unexpected request: %+v", req)
	}
	loaded := loadTaskDB(t, path)
	last := loaded.CommandReceipts[len(loaded.CommandReceipts)-1]
	if last.ApprovalID != "approval:human:1" || !last.RequiresHumanApproval {
		t.Fatalf("claim receipt did not reuse approval: %+v", last)
	}
}

func TestClaimTaskSkipsHumanGatedTaskWithoutApproval(t *testing.T) {
	path := writeTaskDB(t, taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		DecisionGate:        "human-approval-required",
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                    "task-human",
			State:                 task.StateQueued,
			Title:                 "needs approval",
			RecommendedProvider:   "codex",
			RequiresHumanApproval: true,
		}},
	})
	plane := newTestPlane(t, path)

	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req != nil {
		t.Fatalf("human-gated task without approval should not be claimed: %+v", req)
	}
	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-human").State; got != task.StateQueued {
		t.Fatalf("state = %s, want Queued", got)
	}
}

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

func TestCompleteTaskReleasesRuntimeLease(t *testing.T) {
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
	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil || req == nil {
		t.Fatalf("ClaimTask = %+v, %v", req, err)
	}
	reportCtx := contextWithTaskRequest(t, req)
	if err := plane.StartTask(reportCtx, "codex-task"); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.CompleteTask(reportCtx, "codex-task", agentbridge.Result{Status: agentbridge.ResultFailed, Error: "boom"}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	registry := readRuntimeLeaseRegistry(t, plane.leasePath)
	if len(registry.Leases) != 1 || registry.Leases[0].ReleasedAt == nil {
		t.Fatalf("lease should be released after terminal report: %+v", registry.Leases)
	}
}

func TestHeartbeatRefreshesRuntimeLeaseForRunningTask(t *testing.T) {
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
	now := time.Date(2026, 5, 25, 3, 0, 0, 0, time.UTC)
	plane.now = func() time.Time { return now }
	registerRuntimeForProvider(t, plane, "runtime-1", "codex", 2, 0)
	if req, err := plane.ClaimTask(context.Background(), "runtime-1"); err != nil || req == nil {
		t.Fatalf("ClaimTask = %+v, %v", req, err)
	}
	before := readRuntimeLeaseRegistry(t, plane.leasePath).Leases[0].LeaseUntil

	now = now.Add(10 * time.Second)
	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      "runtime-1",
		SlotLimit:      2,
		SlotsInUse:     1,
		RunningTaskIDs: []string{"codex-task"},
	}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	after := readRuntimeLeaseRegistry(t, plane.leasePath).Leases[0].LeaseUntil
	if !after.After(before) {
		t.Fatalf("heartbeat should refresh lease deadline: before=%s after=%s", before, after)
	}
}

func TestStartTaskRejectsExpiredRuntimeLease(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	now := time.Date(2026, 5, 25, 4, 0, 0, 0, time.UTC)
	plane.now = func() time.Time { return now }
	reportCtx := writeRuntimeLease(t, plane, RuntimeLeaseRecord{
		LeaseID:               "runtime-lease:task-1:1",
		TaskID:                "task-1",
		RuntimeID:             "runtime-1",
		CapabilityFingerprint: "runtime-1-fp",
		ClaimedAt:             now.Add(-time.Hour),
		LeaseUntil:            now.Add(-time.Minute),
		FencingToken:          1,
	})

	if err := plane.StartTask(reportCtx, "task-1"); err == nil {
		t.Fatal("StartTask should reject expired runtime lease")
	}
	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-1").State; got != task.StateClaimed {
		t.Fatalf("state should remain Claimed, got %s", got)
	}
}

func TestStartTaskRejectsMismatchedFencingToken(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	reportCtx := writeActiveRuntimeLease(t, plane, "task-1")
	report, ok := controlplane.TaskReportContextFromContext(reportCtx)
	if !ok {
		t.Fatal("missing task report context")
	}
	report.RuntimeFencingToken = 999
	staleCtx := controlplane.ContextWithTaskReport(context.Background(), report)

	if err := plane.StartTask(staleCtx, "task-1"); err == nil {
		t.Fatal("StartTask should reject mismatched fencing token")
	}
	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-1").State; got != task.StateClaimed {
		t.Fatalf("state should remain Claimed, got %s", got)
	}
}

func TestRuntimeRegistryPersistsRegistrationHeartbeatAndDeregister(t *testing.T) {
	path := writeTaskDB(t, taskdb.TaskDB{SchemaVersion: taskdb.TaskDBSchemaVersion})
	plane := newTestPlane(t, path)

	if err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "runtime-1",
		Provider:  "multi",
		Capabilities: map[string]bool{
			"provider.codex.available": true,
		},
		DeviceName: "mac-mini",
	}); err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}
	registry := readRuntimeRegistry(t, plane.registryPath)
	if registry.SchemaVersion != RuntimeRegistrySchemaVersion || registry.TaskDBPath != path {
		t.Fatalf("registry identity mismatch: %+v", registry)
	}
	if len(registry.Runtimes) != 1 || registry.Runtimes[0].RuntimeID != "runtime-1" {
		t.Fatalf("registry runtimes after register: %+v", registry.Runtimes)
	}
	firstHeartbeat := registry.Runtimes[0].LastHeartbeat

	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      "runtime-1",
		SlotLimit:      2,
		SlotsInUse:     1,
		RunningTaskIDs: []string{"task-b", "task-a"},
	}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	registry = readRuntimeRegistry(t, plane.registryPath)
	if len(registry.Runtimes) != 1 || !registry.Runtimes[0].LastHeartbeat.After(firstHeartbeat) {
		t.Fatalf("heartbeat was not persisted: before=%v registry=%+v", firstHeartbeat, registry.Runtimes)
	}
	if registry.Runtimes[0].SlotLimit != 2 || registry.Runtimes[0].SlotsInUse != 1 {
		t.Fatalf("slot heartbeat not persisted: %+v", registry.Runtimes[0])
	}
	if len(registry.Runtimes[0].RunningTaskIDs) != 2 || registry.Runtimes[0].RunningTaskIDs[0] != "task-a" {
		t.Fatalf("running task ids should be sorted: %+v", registry.Runtimes[0].RunningTaskIDs)
	}

	if err := plane.DeregisterRuntime(context.Background(), "runtime-1"); err != nil {
		t.Fatalf("DeregisterRuntime: %v", err)
	}
	registry = readRuntimeRegistry(t, plane.registryPath)
	if len(registry.Runtimes) != 0 {
		t.Fatalf("registry runtimes after deregister: %+v", registry.Runtimes)
	}
}

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

func TestReporterTransitionsRunLifecycleAndTerminalDone(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	reportCtx := writeActiveRuntimeLease(t, plane, "task-1")

	if err := plane.StartTask(reportCtx, "task-1"); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.ReportEvent(reportCtx, "task-1", agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}); err != nil {
		t.Fatalf("ReportEvent: %v", err)
	}
	if err := plane.CompleteTask(reportCtx, "task-1", agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-1").State; got != task.StateValidating {
		t.Fatalf("state = %s, want Validating", got)
	}
	assertTransition(t, db, ir.EventWorkdirPreparing)
	assertTransition(t, db, ir.EventRunStarted)
	assertTransition(t, db, ir.EventRunReportedDone)
}

func TestCompleteTaskSynthesizesRunStartedWhenProviderOmitsLifecycle(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	reportCtx := writeActiveRuntimeLease(t, plane, "task-1")

	if err := plane.StartTask(reportCtx, "task-1"); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.CompleteTask(reportCtx, "task-1", agentbridge.Result{Status: agentbridge.ResultCompleted}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-1").State; got != task.StateValidating {
		t.Fatalf("state = %s, want Validating", got)
	}
	assertTransition(t, db, ir.EventRunStarted)
	assertTransition(t, db, ir.EventRunReportedDone)
}

func TestCompleteTaskCompletedReplayDoesNotAppendDuplicateRunDone(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	reportCtx := writeActiveRuntimeLease(t, plane, "task-1")

	if err := plane.StartTask(reportCtx, "task-1"); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.CompleteTask(reportCtx, "task-1", agentbridge.Result{Status: agentbridge.ResultCompleted}); err != nil {
		t.Fatalf("CompleteTask first: %v", err)
	}
	before := loadTaskDB(t, path)
	if err := plane.CompleteTask(reportCtx, "task-1", agentbridge.Result{Status: agentbridge.ResultCompleted}); err != nil {
		t.Fatalf("CompleteTask replay: %v", err)
	}
	after := loadTaskDB(t, path)
	if len(after.Transitions) != len(before.Transitions) || len(after.CommandReceipts) != len(before.CommandReceipts) {
		t.Fatalf("replay appended mutation: before=%d/%d after=%d/%d", len(before.Transitions), len(before.CommandReceipts), len(after.Transitions), len(after.CommandReceipts))
	}
}

func TestCompleteTaskFailedFromPreparing(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	reportCtx := writeActiveRuntimeLease(t, plane, "task-1")

	if err := plane.StartTask(reportCtx, "task-1"); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.CompleteTask(reportCtx, "task-1", agentbridge.Result{Status: agentbridge.ResultFailed, Error: "boom"}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-1").State; got != task.StateFailed {
		t.Fatalf("state = %s, want Failed", got)
	}
	assertTransition(t, db, ir.EventTaskFailed)
}

func TestCompleteTaskBlockedFromPreparing(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	reportCtx := writeActiveRuntimeLease(t, plane, "task-1")

	if err := plane.StartTask(reportCtx, "task-1"); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.CompleteTask(reportCtx, "task-1", agentbridge.Result{Status: agentbridge.ResultBlocked, Error: "missing required surface"}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-1").State; got != task.StateBlocked {
		t.Fatalf("state = %s, want Blocked", got)
	}
	assertTransition(t, db, ir.EventBlockerRaised)
}

func newTestPlane(t *testing.T, path string) *Plane {
	t.Helper()
	plane, err := New(path)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	now := time.Date(2026, 5, 25, 1, 2, 3, 0, time.UTC)
	plane.now = func() time.Time {
		now = now.Add(time.Second)
		return now
	}
	return plane
}

func registerRuntimeForProvider(t *testing.T, plane *Plane, runtimeID, provider string, slotLimit, slotsInUse int) {
	t.Helper()
	prefix := "provider." + provider + "."
	if err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		RuntimeID:  runtimeID,
		Provider:   "multi",
		SlotLimit:  slotLimit,
		SlotsInUse: slotsInUse,
		Capabilities: map[string]bool{
			prefix + "available":                    true,
			prefix + "supports_streaming":           true,
			prefix + "supports_resume":              true,
			prefix + "supports_system":              true,
			prefix + "supports_max_turns":           true,
			prefix + "supports_mcp":                 true,
			prefix + "supports_tool_hooks":          true,
			prefix + "supports_usage":               true,
			prefix + "supports_worktree":            true,
			prefix + "requires_experimental_opt_in": false,
		},
		CapabilityAttributes: map[string]string{
			prefix + "compatibility_status":   "supported",
			prefix + "capability_fingerprint": runtimeID + "-fp",
		},
	}); err != nil {
		t.Fatalf("RegisterRuntime %s: %v", runtimeID, err)
	}
}

func writeTaskDB(t *testing.T, db taskdb.TaskDB) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "task-db.json")
	if err := taskdb.SaveTaskDB(path, db); err != nil {
		t.Fatalf("SaveTaskDB: %v", err)
	}
	return path
}

func loadTaskDB(t *testing.T, path string) taskdb.TaskDB {
	t.Helper()
	db, err := taskdb.LoadTaskDB(path)
	if err != nil {
		t.Fatalf("LoadTaskDB: %v", err)
	}
	return db
}

func claimedTaskDB() taskdb.TaskDB {
	return taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                  "task-1",
			ProjectID:           "project-1",
			State:               task.StateClaimed,
			Title:               "run it",
			RecommendedProvider: "codex",
		}},
	}
}

func writeActiveRuntimeLease(t *testing.T, plane *Plane, taskID string) context.Context {
	t.Helper()
	now := time.Date(2026, 5, 25, 1, 0, 0, 0, time.UTC)
	return writeRuntimeLease(t, plane, RuntimeLeaseRecord{
		LeaseID:               "runtime-lease:" + taskID + ":1",
		TaskID:                taskID,
		RuntimeID:             "runtime-1",
		CapabilityFingerprint: "runtime-1-fp",
		ClaimedAt:             now,
		LeaseUntil:            now.Add(time.Hour),
		FencingToken:          1,
	})
}

func writeRuntimeLease(t *testing.T, plane *Plane, lease RuntimeLeaseRecord) context.Context {
	t.Helper()
	if err := saveRuntimeLeaseRegistry(plane.leasePath, plane.path, RuntimeLeaseRegistry{Leases: []RuntimeLeaseRecord{lease}}, lease.ClaimedAt); err != nil {
		t.Fatalf("saveRuntimeLeaseRegistry: %v", err)
	}
	return contextWithRuntimeLease(lease)
}

func contextWithTaskRequest(t *testing.T, req *bridge.TaskRequest) context.Context {
	t.Helper()
	report, ok := controlplane.TaskReportContextFromMetadata(req.Metadata)
	if !ok {
		t.Fatalf("request missing task report context metadata: %+v", req.Metadata)
	}
	return controlplane.ContextWithTaskReport(context.Background(), report)
}

func contextWithRuntimeLease(lease RuntimeLeaseRecord) context.Context {
	return controlplane.ContextWithTaskReport(context.Background(), controlplane.TaskReportContext{
		RuntimeLeaseID:               lease.LeaseID,
		RuntimeFencingToken:          lease.FencingToken,
		RuntimeFencingTokenSet:       true,
		RuntimeCapabilityFingerprint: lease.CapabilityFingerprint,
	})
}

func mustFindTask(t *testing.T, db taskdb.TaskDB, id string) taskdb.TaskRecord {
	t.Helper()
	record, ok := findTask(db, id)
	if !ok {
		t.Fatalf("task %s not found", id)
	}
	return record
}

func assertTransition(t *testing.T, db taskdb.TaskDB, event ir.EventType) {
	t.Helper()
	for _, transition := range db.Transitions {
		if transition.EventType == event {
			return
		}
	}
	t.Fatalf("transition %s not found in %+v", event, db.Transitions)
}

func readRuntimeRegistry(t *testing.T, path string) RuntimeRegistry {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read runtime registry: %v", err)
	}
	var registry RuntimeRegistry
	if err := json.Unmarshal(body, &registry); err != nil {
		t.Fatalf("decode runtime registry: %v", err)
	}
	return registry
}

func readRuntimeLeaseRegistry(t *testing.T, path string) RuntimeLeaseRegistry {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read runtime lease registry: %v", err)
	}
	var registry RuntimeLeaseRegistry
	if err := json.Unmarshal(body, &registry); err != nil {
		t.Fatalf("decode runtime lease registry: %v", err)
	}
	return registry
}
