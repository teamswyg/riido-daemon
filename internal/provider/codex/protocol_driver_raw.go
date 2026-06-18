package codex

import (
	"context"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// OnRaw routes a parser RawEvent through either the handshake state
// machine (for "response") or the existing Translate (for notifications
// and server_requests).
func (d *protocolDriver) OnRaw(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	if rawFrameType(raw.Type) == rawFrameResponse {
		return d.handleResponse(ctx, raw, io)
	}
	if rawFrameType(raw.Type) == rawFrameError {
		id, _ := rpcID(raw.Payload)
		if pr, known := d.pending[id]; known {
			delete(d.pending, id)
			return d.failedEvents("codex " + string(pr.method) + " rpc error: " + codexRPCErrorMessage(raw.Payload)), nil, nil
		}
		d.recordRuntimeError(codexRPCErrorMessage(raw.Payload))
		events, cmds, err := Translate(raw)
		d.observeEvents(events)
		return events, cmds, err
	}
	if after, ok := strings.CutPrefix(raw.Type, rawFrameNotificationPrefix); ok {
		if events, handled := d.handleNotification(codexMethod(after), raw); handled {
			return events, nil, nil
		}
	}
	events, cmds, err := Translate(raw)
	d.observeEvents(events)
	return events, cmds, err
}

func (d *protocolDriver) handleNotification(method codexMethod, raw agentbridge.RawEvent) ([]agentbridge.Event, bool) {
	if method == codexMethodError {
		errText := codexNotificationErrorMessage(params(raw))
		d.recordRuntimeError(errText)
		return []agentbridge.Event{{Kind: agentbridge.EventError, Err: errText}}, true
	}
	if method == codexMethodTurnStarted || method == codexMethodTurnStartedSlash {
		d.turnStarted = true
	}
	if method == codexMethodTurnCompleted || method == codexMethodTurnCompleteSlash {
		p := params(raw)
		if d.lastRuntimeError != "" && !d.sawAssistantOutput && stringField(p, "output") == "" {
			return d.failedEvents(d.lastRuntimeError), true
		}
	}
	if method == codexMethodThreadStatusChanged || method == codexMethodThreadStatusAlt {
		events := d.threadStatusEvents(params(raw))
		d.observeEvents(events)
		return events, true
	}
	return nil, false
}
