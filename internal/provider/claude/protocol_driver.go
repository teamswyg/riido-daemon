package claude

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// protocolDriver implements agentbridge.ProtocolDriver for Claude Code's
// `claude -p --input-format stream-json` mode.
type protocolDriver struct {
	req            agentbridge.StartRequest
	written        bool
	streamedBlocks map[int]bool
}

// NewProtocolDriver writes one Claude user-message frame when the session
// starts, then lets the existing stream-json translator handle raw output.
func NewProtocolDriver(req agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return &protocolDriver{req: req}, nil
}

func (d *protocolDriver) OnRaw(
	_ context.Context,
	raw agentbridge.RawEvent,
	_ agentbridge.ProtocolIO,
) ([]agentbridge.Event, []agentbridge.Command, error) {
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

func (d *protocolDriver) OnProcessExit(
	_ context.Context,
	_ agentbridge.ProcessExitStatus,
	_ agentbridge.ProtocolIO,
) ([]agentbridge.Event, error) {
	return nil, nil
}

func (d *protocolDriver) OnClose(_ context.Context, _ agentbridge.ProtocolIO) error {
	return nil
}
