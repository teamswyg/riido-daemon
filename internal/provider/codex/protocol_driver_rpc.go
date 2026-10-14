package codex

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *protocolDriver) sendNotification(ctx context.Context, io agentbridge.ProtocolIO, method codexMethod, params map[string]any) error {
	frame := map[string]any{
		"jsonrpc": "2.0",
		"method":  string(method),
	}
	if params != nil {
		frame["params"] = params
	}
	return d.writeFrame(ctx, io, frame)
}

func (d *protocolDriver) sendRequest(ctx context.Context, io agentbridge.ProtocolIO, method codexMethod, params map[string]any) (int64, error) {
	d.nextID++
	id := d.nextID
	d.pending[id] = pendingRequest{method: method}
	frame := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  string(method),
	}
	if params != nil {
		frame["params"] = params
	}
	return id, d.writeFrame(ctx, io, frame)
}

func (d *protocolDriver) writeFrame(ctx context.Context, io agentbridge.ProtocolIO, frame map[string]any) error {
	b, err := json.Marshal(frame)
	if err != nil {
		return fmt.Errorf("codex driver: marshal %v: %w", frame["method"], err)
	}
	b = append(b, '\n')
	return io.WriteStdin(ctx, b)
}
