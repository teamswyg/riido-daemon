package codex

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexProtocolDriverWritesInitializeOnStart(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	if err := d.OnStart(context.Background(), io); err != nil {
		t.Fatalf("OnStart: %v", err)
	}
	b := io.next(t, time.Second)
	if !strings.Contains(string(b), `"method":"initialize"`) {
		t.Fatalf("first frame is not initialize: %q", b)
	}
	if !strings.Contains(string(b), `"clientInfo"`) {
		t.Fatalf("initialize missing clientInfo: %q", b)
	}
}

func TestCodexProtocolDriverInitializeResponseWritesInitializedAndThreadStart(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{Model: "gpt-5.5"})
	io := newRecordingIO()
	if err := d.OnStart(context.Background(), io); err != nil {
		t.Fatal(err)
	}
	_ = io.next(t, time.Second)

	_, _, err := d.OnRaw(context.Background(), makeResponse(1, map[string]any{"server": "codex"}), io)
	if err != nil {
		t.Fatalf("OnRaw: %v", err)
	}

	first := io.next(t, time.Second)
	if !strings.Contains(string(first), `"method":"initialized"`) {
		t.Fatalf("expected initialized notification, got %q", first)
	}
	second := io.next(t, time.Second)
	if !strings.Contains(string(second), `"method":"thread/start"`) {
		t.Fatalf("expected thread/start, got %q", second)
	}
	if !strings.Contains(string(second), `"model":"gpt-5.5"`) {
		t.Fatalf("thread/start missing model: %q", second)
	}
}
