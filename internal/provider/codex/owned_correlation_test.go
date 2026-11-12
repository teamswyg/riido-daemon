package codex

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestOwnedTurnRejectsStaleAndUnidentifiedTerminals(t *testing.T) {
	for _, p := range []map[string]any{
		{"threadId": "parent", "turn": map[string]any{"id": "older-turn", "status": "completed"}},
		{"threadId": "parent", "turn": map[string]any{"status": "completed"}},
		{"turn": map[string]any{"id": "parent-turn", "status": "completed"}},
		{},
	} {
		d, io := ownedTurnDriver(t)
		events, _, err := d.OnRaw(context.Background(), makeNotification("turn/completed", p), io)
		if err != nil {
			t.Fatal(err)
		}
		for _, event := range events {
			if event.Kind == agentbridge.EventResult {
				t.Fatalf("uncorrelated terminal: %+v", event)
			}
		}
	}
}

func TestOwnedTurnWaitsForCorrelatedReply(t *testing.T) {
	d, io := ownedTurnDriver(t)
	d.owned.id = ""
	d.pending[42] = pendingRequest{method: codexMethodTurnStart}
	for _, id := range []string{"older-turn", "parent-turn"} {
		events, _, err := d.OnRaw(context.Background(), makeNotification("turn/completed", map[string]any{"threadId": "parent", "turn": map[string]any{"id": id, "status": "completed"}}), io)
		if err != nil || len(events) != 0 {
			t.Fatalf("early event was not buffered: %v %+v", err, events)
		}
	}
	events, _, err := d.OnRaw(context.Background(), makeResponse(42, map[string]any{"turn": map[string]any{"id": "parent-turn"}}), io)
	if err != nil {
		t.Fatal(err)
	}
	results := 0
	for _, event := range events {
		if event.Kind == agentbridge.EventResult {
			results++
		}
	}
	if results != 1 || d.owned.id != "parent-turn" || !d.owned.done {
		t.Fatalf("replay=%+v", events)
	}
}

func TestOwnedThreadStatusDoesNotOverrideTurnStatus(t *testing.T) {
	d, io := ownedTurnDriver(t)
	for _, status := range []string{"idle", "completed", "failed", "interrupted"} {
		events, _, _ := d.OnRaw(context.Background(), makeNotification("thread/status/changed", map[string]any{"threadId": "parent", "status": status}), io)
		for _, event := range events {
			if event.Kind == agentbridge.EventResult {
				t.Fatalf("thread status ended owned turn: %+v", event)
			}
		}
	}
}

func TestOwnedTurnProcessExitAfterProgressFails(t *testing.T) {
	d, io := ownedTurnDriver(t)
	_, _, _ = d.OnRaw(context.Background(), makeNotification("item/agentMessage/delta", map[string]any{"threadId": "parent", "turnId": "parent-turn", "itemId": "partial", "delta": "working"}), io)
	events, err := d.OnProcessExit(context.Background(), agentbridge.ProcessExitStatus{Code: 0}, io)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Result.Status != agentbridge.ResultFailed || !strings.Contains(last.Result.Error, "thread=parent turn=parent-turn exit_code=0") {
		t.Fatalf("exit result=%+v", last)
	}
}

func TestOwnedRunPreservesDelegatedToolApproval(t *testing.T) {
	d, io := ownedTurnDriver(t)
	events, _, err := d.OnRaw(context.Background(), makeServerRequest(99, "item/commandExecution/requestApproval", map[string]any{"threadId": "child", "turnId": "child-turn", "itemId": "child-command", "command": "pwd"}), io)
	if err != nil || len(events) != 1 || events[0].Kind != agentbridge.EventToolApprovalNeeded || events[0].Tool.ProviderRequestID != "99" {
		t.Fatalf("delegated approval lost: %v %+v", err, events)
	}
}
