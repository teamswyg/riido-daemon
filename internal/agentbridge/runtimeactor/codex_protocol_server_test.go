package runtimeactor

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

// driveCodexServer simulates the codex app-server: it reads stdin
// frames and emits the matching JSON-RPC responses on stdout.
func driveCodexServer(t *testing.T, r *process.FakeRunning) {
	t.Helper()

	go func() {
		seenInitialize := false
		seenThread := false
		for {
			select {
			case frame, ok := <-r.StdinRecv():
				if !ok {
					return
				}
				if !handleCodexServerFrame(r, string(frame), &seenInitialize, &seenThread) {
					return
				}
			case <-time.After(3 * time.Second):
				return
			}
		}
	}()
}

func handleCodexServerFrame(
	r *process.FakeRunning,
	frame string,
	seenInitialize *bool,
	seenThread *bool,
) bool {
	switch {
	case strings.Contains(frame, `"method":"initialize"`):
		*seenInitialize = true
		r.EmitStdout(jsonRPCResponse(1, map[string]any{"server": "codex-like"}))
	case strings.Contains(frame, `"method":"thread/start"`):
		if !*seenInitialize {
			return false
		}
		*seenThread = true
		r.EmitStdout(jsonRPCResponse(2, map[string]any{"thread_id": "th-1"}))
	case strings.Contains(frame, `"method":"turn/start"`):
		if !*seenThread {
			return false
		}
		r.EmitStdout(jsonRPCResponse(3, map[string]any{}))
		r.EmitStdout(jsonRPCNotification("turn_completed", map[string]any{"output": "all done"}))
	}

	return true
}

func jsonRPCResponse(id int64, result map[string]any) []byte {
	message := map[string]any{"jsonrpc": "2.0", "id": id, "result": result}
	body, _ := json.Marshal(message)
	return append(body, '\n')
}

func jsonRPCNotification(method string, params map[string]any) []byte {
	message := map[string]any{"jsonrpc": "2.0", "method": method, "params": params}
	body, _ := json.Marshal(message)
	return append(body, '\n')
}
