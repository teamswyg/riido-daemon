package runtimeactor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// codexLikeDriver walks the same shape Codex would but with no JSON schema
// dependency: initialize -> thread -> turn -> completed.
type codexLikeDriver struct {
	nextID  int64
	pending map[int64]string
	thread  string
}

func (d *codexLikeDriver) OnStart(ctx context.Context, io agentbridge.ProtocolIO) error {
	return d.send(ctx, io, "initialize", nil)
}

func (d *codexLikeDriver) OnRaw(
	ctx context.Context,
	raw agentbridge.RawEvent,
	io agentbridge.ProtocolIO,
) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "response" {
		return d.onResponse(ctx, raw, io)
	}
	if raw.Type == "notification:turn_completed" {
		return []agentbridge.Event{completedProtocolResult(raw)}, nil, nil
	}
	return nil, nil, nil
}

func (d *codexLikeDriver) OnProcessExit(
	_ context.Context,
	_ agentbridge.ProcessExitStatus,
	_ agentbridge.ProtocolIO,
) ([]agentbridge.Event, error) {
	return nil, nil
}

func (d *codexLikeDriver) OnClose(_ context.Context, _ agentbridge.ProtocolIO) error { return nil }

func (d *codexLikeDriver) onResponse(
	ctx context.Context,
	raw agentbridge.RawEvent,
	io agentbridge.ProtocolIO,
) ([]agentbridge.Event, []agentbridge.Command, error) {
	id, _ := payloadInt64(raw.Payload, "id")
	method := d.pending[id]
	delete(d.pending, id)
	switch method {
	case "initialize":
		return nil, nil, d.send(ctx, io, "thread/start", nil)
	case "thread/start":
		result, _ := raw.Payload["result"].(map[string]any)
		d.thread, _ = result["thread_id"].(string)
		return nil, nil, d.send(ctx, io, "turn/start", map[string]any{"thread_id": d.thread})
	case "turn/start":
		return []agentbridge.Event{{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}}, nil, nil
	default:
		return nil, nil, nil
	}
}
