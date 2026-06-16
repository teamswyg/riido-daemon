package riidoapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestServerRejectsIllegalTransition(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response TransitionResponse
	err := client.Request(context.Background(), "transition", TransitionRequest{
		TaskID:    "task:test",
		ToState:   string(task.StateRunning),
		EventType: string(ir.EventRunStarted),
	}, &response)
	if err == nil {
		t.Fatal("expected illegal transition error")
	}
}

func TestServerRejectsMissingApproval(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response TransitionResponse
	err := client.Request(context.Background(), "transition", TransitionRequest{
		TaskID:    "task:test",
		ToState:   string(task.StateQueued),
		EventType: string(ir.EventTaskQueued),
		Actor:     "human",
		Source:    "test",
	}, &response)
	if err == nil {
		t.Fatal("expected missing approval_id to be rejected")
	}
}

func serveTestAPI(t *testing.T) (string, string, func()) {
	t.Helper()
	return serveTestAPIWithState(t, task.StateCreated)
}

func serveTestAPIWithState(t *testing.T, state task.TaskState) (string, string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "rapi-")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	socketPath := filepath.Join(dir, "riido.sock")
	taskDBPath := filepath.Join(dir, "task-db.json")
	db := taskdb.EmptyTaskDB()
	db.UpdatedAt = "2026-05-20T08:00:00Z"
	db.RecommendedProvider = "codex"
	db.RecommendedDecisionLLM = "codex"
	db.DecisionGate = "human-approval-required"
	db.ProviderCandidates = []taskdb.ProviderCandidate{
		{ID: "codex", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
	}
	db.Tasks = []taskdb.TaskRecord{
		{
			ID:                     "task:test",
			ProjectID:              "macmini-workspace",
			State:                  state,
			SourceDocumentID:       "mws.test",
			SourceDocumentPath:     "docs/TEST.md",
			Title:                  "테스트",
			Owner:                  "local",
			SourceStatus:           "seed",
			RecommendedProvider:    "codex",
			RecommendedDecisionLLM: "codex",
			RequiresHumanApproval:  true,
			HarnessNextDirection:   "top-down",
			CreatedAt:              "2026-05-20T08:00:00Z",
			UpdatedAt:              "2026-05-20T08:00:00Z",
			TransitionCount:        1,
		},
	}
	db.Transitions = []taskdb.TaskTransitionRecord{
		{
			ID:         "transition:task:test:created",
			TaskID:     "task:test",
			ToState:    task.StateCreated,
			EventType:  ir.EventTaskCreated,
			Actor:      "riido",
			Source:     "test",
			RecordedAt: "2026-05-20T08:00:00Z",
		},
	}
	if err := taskdb.SaveTaskDB(taskDBPath, db); err != nil {
		t.Fatalf("SaveTaskDB failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan error, 1)
	go func() {
		errs <- NewServer(Config{SocketPath: socketPath, TaskDBPath: taskDBPath}).Serve(ctx)
	}()
	waitForAPI(t, socketPath)
	return socketPath, taskDBPath, func() {
		cancel()
		select {
		case err := <-errs:
			if err != nil {
				t.Fatalf("Serve returned error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for API server shutdown")
		}
	}
}

func waitForAPI(t *testing.T, socketPath string) {
	t.Helper()
	client := NewClient(socketPath)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var status Status
		if err := client.Request(context.Background(), "status", nil, &status); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("riido API did not become ready at %s", socketPath)
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
