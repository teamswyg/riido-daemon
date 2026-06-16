package riidoapi

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/task"
)

func TestServerValidateUsesCallerContext(t *testing.T) {
	_, taskDBPath, stop := serveTestAPIWithState(t, task.StateValidating)
	defer stop()

	params, err := json.Marshal(ValidateRequest{
		TaskID:         "task:test",
		Command:        "sleep 2",
		TimeoutSeconds: 30,
		ApprovalID:     "approval:test:validate-cancel",
		CommandID:      "command:test:validate-cancel",
	})
	if err != nil {
		t.Fatalf("marshal validate params: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	started := time.Now()
	response, err := NewServer(Config{TaskDBPath: taskDBPath}).validateTask(ctx, params)
	if err != nil {
		t.Fatalf("validateTask returned error: %v", err)
	}
	if elapsed := time.Since(started); elapsed > 500*time.Millisecond {
		t.Fatalf("validateTask ignored caller context: elapsed=%s", elapsed)
	}
	if response.Validation.Result != "failed" || response.Validation.ExitCode == 0 {
		t.Fatalf("canceled validation should be recorded as failed: %#v", response.Validation)
	}
}

func TestServerValidateRequiresApprovalBeforeExecution(t *testing.T) {
	socketPath, _, stop := serveTestAPIWithState(t, task.StateValidating)
	defer stop()

	marker := filepath.Join(t.TempDir(), "should-not-exist")
	client := NewClient(socketPath)
	var response ValidateResponse
	err := client.Request(context.Background(), "validate", ValidateRequest{
		TaskID:         "task:test",
		Command:        "touch " + marker,
		Workdir:        t.TempDir(),
		TimeoutSeconds: 5,
		CommandID:      "command:test:validate",
		Provider:       "codex",
		DecisionLLM:    "codex",
	}, &response)
	if err == nil {
		t.Fatal("expected missing approval_id to be rejected")
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("validation command should not execute without approval, statErr=%v", statErr)
	}
}
