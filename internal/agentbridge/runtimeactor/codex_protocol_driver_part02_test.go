package runtimeactor

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// TestRuntimeActorRunsCodexWithProtocolDriver is the M-9 integration
// success criterion: a TaskRequest goes through RuntimeActor.Submit →
// SessionActor → ProtocolDriver, walks the full handshake against a
// fake process, and reaches ResultCompleted WITHOUT any inline test
// driver writing to stdin.
func TestRuntimeActorRunsCodexWithProtocolDriver(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters:      []agentbridge.Adapter{codexLikeAdapter{}},
		MaxConcurrent: 1,
	})

	h, err := a.Submit(context.Background(), bridge.TaskRequest{
		ID: "t-codex", Provider: "codex-like", Prompt: "do the thing",
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	running := waitForRunning(t, p, 0, time.Second)

	// Drain events so the session loop doesn't block.
	go func() {
		for range h.Events() {
		}
	}()

	// "Codex server" side. It must observe stdin writes and emit the
	// matching JSON-RPC responses on stdout.
	driveCodexServer(t, running)

	// The session must reach ResultCompleted purely through the driver
	// path — no inline test-driver-side stdin writes here.
	select {
	case res := <-h.Result():
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("status: %s", res.Status)
		}
		if res.Output != "all done" {
			t.Fatalf("output: %q", res.Output)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("session never reached ResultCompleted via ProtocolDriver")
	}

	// Slot must release.
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		s, _ := a.Status(context.Background())
		if s.RunningSessions == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("RunningSessions never returned to 0")
}

// driveCodexServer simulates the codex app-server: it reads stdin
// frames and emits the matching JSON-RPC responses on stdout.
func driveCodexServer(t *testing.T, r *process.FakeRunning) {
	t.Helper()
	go func() {
		seenInitialize := false
		seenThread := false
		seenTurn := false
		for {
			select {
			case frame, ok := <-r.StdinRecv():
				if !ok {
					return
				}
				s := string(frame)
				switch {
				case strings.Contains(s, `"method":"initialize"`):
					seenInitialize = true
					r.EmitStdout(jsonRPCResponse(1, map[string]any{"server": "codex-like"}))
				case strings.Contains(s, `"method":"thread/start"`):
					if !seenInitialize {
						return
					}
					seenThread = true
					r.EmitStdout(jsonRPCResponse(2, map[string]any{"thread_id": "th-1"}))
				case strings.Contains(s, `"method":"turn/start"`):
					if !seenThread {
						return
					}
					seenTurn = true
					r.EmitStdout(jsonRPCResponse(3, map[string]any{}))
					// Then emit turn_completed notification.
					r.EmitStdout(jsonRPCNotification("turn_completed", map[string]any{"output": "all done"}))
				default:
					_ = seenTurn // unused in failure paths
				}
			case <-time.After(3 * time.Second):
				return
			}
		}
	}()
}

func jsonRPCResponse(id int64, result map[string]any) []byte {
	m := map[string]any{"jsonrpc": "2.0", "id": id, "result": result}
	b, _ := json.Marshal(m)
	return append(b, '\n')
}

func jsonRPCNotification(method string, params map[string]any) []byte {
	m := map[string]any{"jsonrpc": "2.0", "method": method, "params": params}
	b, _ := json.Marshal(m)
	return append(b, '\n')
}

// Compile-time guarantee: the test-only adapter satisfies the optional
// interface RuntimeActor will probe for.
var _ agentbridge.ProtocolDriverProvider = codexLikeAdapter{}
