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
