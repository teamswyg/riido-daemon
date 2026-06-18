package riidoapi

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

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
