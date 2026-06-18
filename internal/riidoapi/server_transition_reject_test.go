package riidoapi

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
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
