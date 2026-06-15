package riidoapi

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestServerExposesTaskDB(t *testing.T) {
	socketPath, taskDBPath, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var status Status
	if err := client.Request(context.Background(), "status", nil, &status); err != nil {
		t.Fatalf("status request failed: %v", err)
	}
	if status.SchemaVersion != StatusSchemaVersion {
		t.Fatalf("unexpected status schema: %s", status.SchemaVersion)
	}
	if status.Transport != string(LocalTransportUnixSocket) {
		t.Fatalf("unexpected API transport: %s", status.Transport)
	}
	if status.TaskCount != 1 {
		t.Fatalf("unexpected task count: %d", status.TaskCount)
	}
	if status.EvidenceCount != 0 {
		t.Fatalf("unexpected evidence count: %d", status.EvidenceCount)
	}
	if status.CommandReceiptCount != 0 {
		t.Fatalf("unexpected command receipt count: %d", status.CommandReceiptCount)
	}
	if status.TaskDBPath != taskDBPath {
		t.Fatalf("unexpected task DB path: %s", status.TaskDBPath)
	}

	var db taskdb.TaskDB
	if err := client.Request(context.Background(), "tasks", nil, &db); err != nil {
		t.Fatalf("tasks request failed: %v", err)
	}
	if db.SchemaVersion != taskdb.TaskDBSchemaVersion {
		t.Fatalf("unexpected task DB schema: %s", db.SchemaVersion)
	}
	if db.Tasks[0].ID != "task:test" {
		t.Fatalf("unexpected task: %#v", db.Tasks[0])
	}
	if db.Tasks[0].RecommendedProvider != "codex" {
		t.Fatalf("unexpected task recommended provider: %s", db.Tasks[0].RecommendedProvider)
	}
	if !db.Tasks[0].RequiresHumanApproval {
		t.Fatalf("task queue should expose human approval gate")
	}
}

func TestServerAppliesTransition(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response TransitionResponse
	err := client.Request(context.Background(), "transition", TransitionRequest{
		TaskID:     "task:test",
		ToState:    string(task.StateQueued),
		EventType:  string(ir.EventTaskQueued),
		Actor:      "human",
		Source:     "test",
		Reason:     "ready",
		ApprovalID: "approval:test:transition",
		CommandID:  "command:test:transition",
	}, &response)
	if err != nil {
		t.Fatalf("transition request failed: %v", err)
	}
	if response.Task.State != task.StateQueued {
		t.Fatalf("unexpected task state: %s", response.Task.State)
	}
	if response.Transition.EventType != ir.EventTaskQueued {
		t.Fatalf("unexpected transition: %#v", response.Transition)
	}
	if response.Receipt.ID == "" || response.Receipt.ID != response.Transition.CommandReceiptID {
		t.Fatalf("transition should return command receipt: %#v", response)
	}
}

func TestServerAddsEvidence(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response EvidenceResponse
	err := client.Request(context.Background(), "evidence", EvidenceRequest{
		TaskID:            "task:test",
		Command:           "cargo check",
		ExitCode:          0,
		Actor:             "daemon",
		Source:            "test",
		Summary:           "build passed",
		ApprovalID:        "approval:test:evidence",
		CommandID:         "command:test:evidence",
		ValidationGate:    "validation:api:cargo-check",
		ProviderRunID:     "provider-run:api:evidence",
		ProviderRunResult: "passed",
	}, &response)
	if err != nil {
		t.Fatalf("evidence request failed: %v", err)
	}
	if response.Evidence.Result != "passed" {
		t.Fatalf("unexpected evidence result: %s", response.Evidence.Result)
	}
	if response.Evidence.DocumentID != "mws.test" {
		t.Fatalf("unexpected evidence document id: %s", response.Evidence.DocumentID)
	}
	if response.Task.EvidenceCount != 1 {
		t.Fatalf("unexpected task evidence count: %d", response.Task.EvidenceCount)
	}
	if response.Evidence.ValidationGate != "validation:api:cargo-check" {
		t.Fatalf("unexpected validation gate: %s", response.Evidence.ValidationGate)
	}
	if response.Evidence.ProviderRunID != "provider-run:api:evidence" || response.Evidence.ProviderRunResult != "passed" {
		t.Fatalf("unexpected provider run evidence: %#v", response.Evidence)
	}
	if response.Receipt.ID == "" || response.Receipt.ID != response.Evidence.CommandReceiptID {
		t.Fatalf("evidence should return command receipt: %#v", response)
	}
}

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
