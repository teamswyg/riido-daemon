package codex

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// --- B4: streaming agent_message → TextDelta ---

func TestCodexProtocolDriverStreamingNotificationTranslatesToEvents(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)

	events, _, err := d.OnRaw(context.Background(), makeNotification("agent_message", map[string]any{"text": "hi"}), io)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != agentbridge.EventTextDelta || events[0].Text != "hi" {
		t.Fatalf("text delta: %+v", events)
	}
}

// --- B5: approve_command → ToolApprovalNeeded ---

func TestCodexProtocolDriverApprovalRequestTranslatesToApprovalEvent(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)

	events, _, err := d.OnRaw(
		context.Background(),
		makeServerRequest(42, "approve_command", map[string]any{"id": "cmd-1", "command": "ls"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != agentbridge.EventToolApprovalNeeded {
		t.Fatalf("approval: %+v", events)
	}
	if events[0].Tool.ID != "cmd-1" {
		t.Fatalf("tool ref: %+v", events[0].Tool)
	}
}

// --- B6: turn_completed → ResultCompleted ---

func TestCodexProtocolDriverTurnCompletedReturnsResult(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)

	events, _, err := d.OnRaw(
		context.Background(),
		makeNotification("turn_completed", map[string]any{"output": "all done"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("no events")
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", last)
	}
}

func TestCodexProtocolDriverRPCErrorFailsPendingRequest(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second) // initialize pending id=1

	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      int64(1),
		"error":   map[string]any{"message": "missing auth"},
	}
	events, _, err := d.OnRaw(context.Background(), agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    "error",
		Payload: payload,
	}, io)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("expected failed result for rpc error, got %+v", events)
	}
	if !strings.Contains(last.Result.Error, "initialize") || !strings.Contains(last.Result.Error, "missing auth") {
		t.Fatalf("missing rpc error context: %+v", last.Result)
	}
}

func TestCodexProtocolDriverErrorNotificationThenEmptyCompletionFails(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)

	events, _, err := d.OnRaw(
		context.Background(),
		makeNotification("error", map[string]any{"message": "failed to connect to websocket: 401 Unauthorized"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != agentbridge.EventError {
		t.Fatalf("expected non-terminal error event first, got %+v", events)
	}

	events, _, err = d.OnRaw(
		context.Background(),
		makeNotification("turn_completed", map[string]any{"output": ""}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("expected failed result after empty completion, got %+v", events)
	}
	if !strings.Contains(last.Result.Error, "401 Unauthorized") {
		t.Fatalf("expected retained runtime error, got %+v", last.Result)
	}
}

func TestCodexProtocolDriverErrorNotificationCanRecoverWithText(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)

	_, _, _ = d.OnRaw(
		context.Background(),
		makeNotification("error", map[string]any{"message": "temporary stream disconnect"}),
		io,
	)
	events, _, err := d.OnRaw(
		context.Background(),
		makeNotification("agent_message", map[string]any{"text": "작업 완료"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != agentbridge.EventTextDelta {
		t.Fatalf("expected text delta, got %+v", events)
	}

	events, _, err = d.OnRaw(
		context.Background(),
		makeNotification("turn_completed", map[string]any{"output": ""}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("expected recovered completion, got %+v", events)
	}
}
