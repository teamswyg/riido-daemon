package riidoapi

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func TestServerValidatesTaskAndTransitionsFromValidating(t *testing.T) {
	socketPath, _, stop := serveTestAPIWithState(t, task.StateValidating)
	defer stop()

	client := NewClient(socketPath)
	var response ValidateResponse
	err := client.Request(context.Background(), "validate", ValidateRequest{
		TaskID:         "task:test",
		Command:        "printf ok",
		Workdir:        t.TempDir(),
		TimeoutSeconds: 5,
		Actor:          "daemon",
		Source:         "test",
		ApprovalID:     "approval:test:validate",
		CommandID:      "command:test:validate",
		Provider:       "codex",
		DecisionLLM:    "codex",
	}, &response)
	if err != nil {
		t.Fatalf("validate request failed: %v", err)
	}
	assertValidationResponse(t, response)
}

func assertValidationResponse(t *testing.T, response ValidateResponse) {
	t.Helper()
	if response.Validation.Result != "passed" {
		t.Fatalf("unexpected validation result: %#v", response.Validation)
	}
	if response.Evidence.Result != "passed" || response.Evidence.CommandReceiptID != response.Receipt.ID {
		t.Fatalf("unexpected evidence/receipt pair: %#v", response)
	}
	if response.Task.State != task.StatePatchReady {
		t.Fatalf("expected task to transition to PatchReady, got %s", response.Task.State)
	}
	if response.Transition == nil || response.Transition.EventType != ir.EventValidationPassed {
		t.Fatalf("expected validation passed transition: %#v", response.Transition)
	}
	if response.TransitionReceipt == nil || response.TransitionReceipt.CommandID != "command:test:validate:transition" {
		t.Fatalf("expected transition receipt pair: %#v", response.TransitionReceipt)
	}
}
