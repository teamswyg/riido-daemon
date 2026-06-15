package taskdbplane

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

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
