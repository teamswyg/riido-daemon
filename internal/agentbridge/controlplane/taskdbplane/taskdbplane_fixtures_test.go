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
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

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
