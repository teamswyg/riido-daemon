package codex

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// --- handshake state machine (NOT a RunState FSM — just transport progress) ---

func (d *protocolDriver) handleResponse(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	id, ok := rpcID(raw.Payload)
	if !ok {
		return nil, nil, nil
	}
	pr, known := d.pending[id]
	if !known {
		// Stray response — no-op.
		return nil, nil, nil
	}
	delete(d.pending, id)

	switch pr.method {
	case codexMethodInitialize:
		return d.handleInitializeResponse(ctx, io)
	case codexMethodThreadStart, codexMethodThreadResume:
		return d.handleThreadReadyResponse(ctx, raw, io)
	case codexMethodTurnStart:
		return d.handleTurnStartResponse()
	default:
		return nil, nil, nil
	}
}
