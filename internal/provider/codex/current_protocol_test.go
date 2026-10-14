package codex

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexModernApproval(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	es, _, e := d.OnRaw(context.Background(), makeServerRequest(42, "item/commandExecution/requestApproval", map[string]any{"threadId": "t", "turnId": "u", "itemId": "cmd", "command": "pwd"}), newRecordingIO())
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range es {
		if v.Kind == agentbridge.EventToolApprovalNeeded {
			return
		}
	}
	t.Fatalf("approval request becomes %+v", es)
}

func TestCodexFailedTurnAfterOutput(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	ctx := context.Background()
	io := newRecordingIO()
	_, _, _ = d.OnRaw(ctx, makeNotification("item/agentMessage/delta", map[string]any{"delta": "partial answer"}), io)
	_, _, _ = d.OnRaw(ctx, makeNotification("error", map[string]any{"error": map[string]any{"message": "quota exceeded"}, "willRetry": false}), io)
	es, _, e := d.OnRaw(ctx, makeNotification("turn/completed", map[string]any{"threadId": "t", "turn": map[string]any{"id": "u", "status": "failed", "error": map[string]any{"message": "quota exceeded"}}}), io)
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range es {
		if v.Kind == agentbridge.EventResult {
			if v.Result.Status != agentbridge.ResultFailed || v.Result.Error != "quota exceeded" {
				t.Fatalf("failed turn classified as %+v", v.Result)
			}
			return
		}
	}
	t.Fatal("no terminal result")
}

func TestModernApprovalRoundTripPreservesID(t *testing.T) {
	for _, id := range []any{float64(42), "request-42"} {
		raw := agentbridge.RawEvent{Type: "server_request:item/commandExecution/requestApproval", Payload: map[string]any{"id": id, "params": map[string]any{"itemId": "cmd", "command": "pwd"}}}
		events, _, err := Translate(raw)
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 || events[0].Tool.ID != "cmd" {
			t.Fatalf("events=%+v", events)
		}
		b, err := BuildProviderInput(agentbridge.Command{Kind: agentbridge.CommandApproveTool, ProviderRequestID: events[0].Tool.ProviderRequestID})
		if err != nil {
			t.Fatal(err)
		}
		var response map[string]any
		if err = json.Unmarshal(b, &response); err != nil {
			t.Fatal(err)
		}
		if response["id"] != id || response["result"].(map[string]any)["decision"] != "accept" {
			t.Fatalf("response=%s", b)
		}
	}
}

func TestIdleAfterPartialOutputDoesNotCompleteTurn(t *testing.T) {
	d := &protocolDriver{turnStarted: true, sawAssistantOutput: true}
	for _, event := range d.threadStatusEvents(map[string]any{"status": map[string]any{"type": "idle"}}) {
		if event.Kind == agentbridge.EventResult {
			t.Fatal("idle ended turn before terminal result")
		}
	}
	events := translateNotification("turn/completed", map[string]any{"turn": map[string]any{"status": "interrupted"}})
	if events[0].Result.Status != agentbridge.ResultCancelled {
		t.Fatalf("events=%+v", events)
	}
}
