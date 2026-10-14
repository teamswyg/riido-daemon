package codex

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexNamesThreadAfterFirstTurnStarts(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{Prompt: "Fix bug", Metadata: map[string]string{"task_title": "버그 수정"}})
	io := newRecordingIO()
	ctx := context.Background()
	if err := d.OnStart(ctx, io); err != nil {
		t.Fatal(err)
	}
	_ = io.next(t, time.Second)
	_, _, _ = d.OnRaw(ctx, makeResponse(1, nil), io)
	_ = io.next(t, time.Second)
	_ = io.next(t, time.Second)
	_, _, _ = d.OnRaw(ctx, makeResponse(2, map[string]any{"thread": map[string]any{"id": "native"}}), io)
	turn := io.next(t, time.Second)
	var f map[string]any
	if err := json.Unmarshal(turn, &f); err != nil {
		t.Fatal(err)
	}
	if f["method"] != "turn/start" {
		t.Fatalf("named before persisted turn: %s", turn)
	}
	_, _, _ = d.OnRaw(ctx, makeResponse(3, nil), io)
	name := io.next(t, time.Second)
	if err := json.Unmarshal(name, &f); err != nil {
		t.Fatal(err)
	}
	p := f["params"].(map[string]any)
	if f["method"] != "thread/name/set" || p["threadId"] != "native" || p["name"] != "버그 수정" {
		t.Fatalf("incorrect task name: %s", name)
	}
	events, _, err := d.OnRaw(ctx, agentbridge.RawEvent{Type: "error", Payload: map[string]any{"id": int64(4), "error": map[string]any{"message": "unknown method"}}}, io)
	if err != nil || len(events) != 1 || events[0].Kind != agentbridge.EventWarning {
		t.Fatalf("optional naming failed execution: %+v %v", events, err)
	}
}
