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
	req            agentbridge.StartRequest
	written        bool // OnStart should run only once even if reinvoked
	streamedBlocks map[int]bool
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
	if raw.Type == "stream_event" {
		event, _ := raw.Payload["event"].(map[string]any)
		if event["type"] == "message_start" {
			d.streamedBlocks = nil
		}
		delta, _ := event["delta"].(map[string]any)
		if event["type"] == "content_block_delta" && (delta["type"] == "text_delta" || delta["type"] == "thinking_delta") {
			if d.streamedBlocks == nil {
				d.streamedBlocks = map[int]bool{}
			}
			index, _ := event["index"].(float64)
			d.streamedBlocks[int(index)] = true
		}
	}
	if raw.Type == "assistant" && len(d.streamedBlocks) > 0 {
		// Copy the payload: parser data may also be consumed by diagnostics.
		message, _ := raw.Payload["message"].(map[string]any)
		content, _ := message["content"].([]any)
		filtered := make([]any, 0, len(content))
		for index, item := range content {
			block, _ := item.(map[string]any)
			if d.streamedBlocks[index] && (block["type"] == "text" || block["type"] == "thinking") {
				continue
			}
			filtered = append(filtered, item)
		}
		raw.Payload = map[string]any{"message": map[string]any{"content": filtered}}
		d.streamedBlocks = nil
	}
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
