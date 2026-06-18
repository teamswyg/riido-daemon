package taskvalidation

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func assertValidationPassedRun(t *testing.T, result Result, commandID string) {
	t.Helper()
	if result.Validation.Result != "passed" || result.Evidence.Result != "passed" {
		t.Fatalf("validation should pass: %#v", result)
	}
	if result.Task.State != task.StatePatchReady {
		t.Fatalf("validating task should move to patch-ready: %s", result.Task.State)
	}
	if result.Transition == nil || result.Transition.EventType != ir.EventValidationPassed {
		t.Fatalf("missing validation transition: %#v", result.Transition)
	}
	assertValidationReceipts(t, result, commandID)
}

func assertValidationReceipts(t *testing.T, result Result, commandID string) {
	t.Helper()
	if result.Receipt.CommandID != commandID {
		t.Fatalf("evidence receipt command id mismatch: %#v", result.Receipt)
	}
	if result.TransitionReceipt == nil || result.TransitionReceipt.CommandID != commandID+":transition" {
		t.Fatalf("transition receipt command id mismatch: %#v", result.TransitionReceipt)
	}
}
