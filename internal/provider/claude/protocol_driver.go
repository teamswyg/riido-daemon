package claude

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// protocolDriver implements agentbridge.ProtocolDriver for Claude Code's
// `claude -p --input-format stream-json` mode.
//
// Claude reads ONE stream-json user-message frame from stdin and then
// waits for EOF before responding. The Codex driver is long-lived
// JSON-RPC; the Claude driver is simpler:
//
//	OnStart            → write {"type":"user","message":{...}} to stdin
//	                   → CloseStdin (EOF signal)
//	OnRaw              → delegate to claude.Translate (existing logic)
//	OnProcessExit      → no-op (no pending requests to fail; Claude does
//	                      not use server-initiated RPC requests)
//	OnClose            → no-op
//
// Discovered during real-CLI smoke (audit "real CLI integration matrix"
// op-quality work): without this driver the session actor never wrote
// the prompt to stdin and `claude -p` blocked on stdin indefinitely,
// hitting the hard-timeout.
type protocolDriver struct {
	req     agentbridge.StartRequest
	written bool // OnStart should run only once even if reinvoked
}

// NewProtocolDriver returns a provider-neutral ProtocolDriver that writes the
// Claude user-message stream-json frame on session start.
func NewProtocolDriver(req agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return &protocolDriver{req: req}, nil
}

// OnStart writes the user-message stream-json frame to claude's stdin
// and immediately closes stdin so claude knows there's no more input.
// Claude then processes the prompt and streams its response back on
// stdout, which the session actor feeds through Translate the normal
// way.
func (d *protocolDriver) OnStart(ctx context.Context, io agentbridge.ProtocolIO) error {
	if d.written {
		return nil
	}
	frame := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role": "user",
			"content": []map[string]any{
				{"type": "text", "text": d.req.Prompt},
			},
		},
	}
	body, err := json.Marshal(frame)
	if err != nil {
		return fmt.Errorf("claude driver: marshal user frame: %w", err)
	}
	body = append(body, '\n')
	if err := io.WriteStdin(ctx, body); err != nil {
		return fmt.Errorf("claude driver: write user frame: %w", err)
	}
	if err := io.CloseStdin(ctx); err != nil {
		return fmt.Errorf("claude driver: close stdin: %w", err)
	}
	d.written = true
	return nil
}

// OnRaw delegates to the existing Translate. The driver does not need
// to inspect raw events itself — Claude's stream-json semantics are
// already covered by the pure-translator path.
func (d *protocolDriver) OnRaw(_ context.Context, raw agentbridge.RawEvent, _ agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	return Translate(raw)
}

// OnProcessExit is a no-op for Claude — there are no pending RPC
// requests to fail. The canonical reducer's process-exit policy
// handles "exited without provider result".
func (d *protocolDriver) OnProcessExit(_ context.Context, _ agentbridge.ProcessExitStatus, _ agentbridge.ProtocolIO) ([]agentbridge.Event, error) {
	return nil, nil
}

// OnClose is a no-op for Claude.
func (d *protocolDriver) OnClose(_ context.Context, _ agentbridge.ProtocolIO) error {
	return nil
}
