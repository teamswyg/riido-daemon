package codex

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestCodexHandshakeStdinHistory(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	proc, _ := fake.Start(context.Background(), process.Command{Executable: "codex"})
	running := proc.(*process.FakeRunning)

	mustWriteJSONRPC(t, running, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize"})
	mustWriteJSONRPC(t, running, map[string]any{"jsonrpc": "2.0", "method": "initialized"})
	mustWriteJSONRPC(t, running, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "thread/start"})
	mustWriteJSONRPC(t, running, map[string]any{"jsonrpc": "2.0", "id": 3, "method": "turn/start"})

	assertCodexHandshakeStdinHistory(t, running)
}

func assertCodexHandshakeStdinHistory(t *testing.T, running *process.FakeRunning) {
	t.Helper()
	want := []string{"initialize", "initialized", "thread/start", "turn/start"}
	for i, w := range want {
		select {
		case got := <-running.StdinRecv():
			if !strings.Contains(string(got), w) {
				t.Fatalf("stdin frame #%d: want %q got %q", i, w, got)
			}
		case <-time.After(time.Second):
			t.Fatalf("stdin frame #%d (%s) never written", i, w)
		}
	}
}
