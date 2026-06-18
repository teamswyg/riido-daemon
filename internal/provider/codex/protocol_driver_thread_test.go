package codex

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexProtocolDriverThreadResponseWritesTurnStart(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{Prompt: "do the thing", Model: "gpt-5.5"})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)
	_, _, _ = d.OnRaw(context.Background(), makeResponse(1, nil), io)
	_ = io.next(t, time.Second)
	_ = io.next(t, time.Second)

	_, _, _ = d.OnRaw(context.Background(), makeResponse(2, map[string]any{"thread": map[string]any{"id": "th-xyz"}}), io)

	frame := io.next(t, time.Second)
	if !strings.Contains(string(frame), `"method":"turn/start"`) {
		t.Fatalf("expected turn/start, got %q", frame)
	}
	for _, want := range []string{"th-xyz", "do the thing", `"input"`, `"model":"gpt-5.5"`} {
		if !strings.Contains(string(frame), want) {
			t.Fatalf("turn/start missing %s: %q", want, frame)
		}
	}
}
