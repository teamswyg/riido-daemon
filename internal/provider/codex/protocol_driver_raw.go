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
	events, commands, err := d.onRaw(ctx, raw, io)
	for i := range events {
		if events[i].Kind == agentbridge.EventLog {
			events[i].LocalOnly = true
		}
	}
	return events, commands, err
}

func (d *protocolDriver) onRaw(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	if rawFrameType(raw.Type) == rawFrameResponse {
		return d.handleResponse(ctx, raw, io)
	}
	if rawFrameType(raw.Type) == rawFrameError {
		id, _ := rpcID(raw.Payload)
		if pr, known := d.pending[id]; known {
			delete(d.pending, id)
			if pr.method == codexMethodQueueDelete && d.shared != nil && d.shared.startError != "" {
				d.shared.done = true
				return d.failedEvents(d.shared.startError + ". 대기 중인 요청을 정리하지 못했습니다. 재요청 전에 Codex 앱의 대기열을 확인해주세요"), nil, nil
			}
			if pr.method == codexMethodThreadResume && d.req.ResumeSessionID != "" && codexRPCErrorMessage(raw.Payload) == "thread "+d.req.ResumeSessionID+" already has an active writer" {
				return d.attachSharedWriter(ctx, io)
			}
			if pr.method == codexMethodThreadNameSet {
				return []agentbridge.Event{{Kind: agentbridge.EventWarning, Err: "Codex thread name unavailable: " + codexRPCErrorMessage(raw.Payload)}}, nil, nil
			}
			return codexFailureEvents(raw.Payload, "codex "+string(pr.method)+" rpc error: "+codexRPCErrorMessage(raw.Payload)), nil, nil
		}
		d.recordRuntimeError(codexRPCErrorMessage(raw.Payload))
		events, cmds, err := Translate(raw)
		d.observeEvents(events)
		return events, cmds, err
	}
	if d.shared != nil {
		// The child does not own the native turn. Only correlated read replies
		// may publish its output or completion.
		return nil, nil, nil
	}
	if raw.Type == rawFrameServerRequestPrefix+codexMCPApprovalMethod {
		return d.handleMCPApproval(ctx, raw, io)
	}
	if after, ok := strings.CutPrefix(raw.Type, rawFrameNotificationPrefix); ok {
		return d.notificationEvents(codexMethod(after), raw)
	}
	events, cmds, err := Translate(raw)
	events = d.ownedResultEvents(raw, events)
	d.observeEvents(events)
	return events, cmds, err
}

// notificationEvents has no protocol I/O, so idle recovery can consume owned
// output and terminal evidence without dispatching requests or publishing logs.
func (d *protocolDriver) notificationEvents(method codexMethod, raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	if events, handled := d.guardOwnedNotification(raw); handled {
		return events, nil, nil
	}
	if events, handled := d.handleNotification(method, raw); handled {
		return d.ownedResultEvents(raw, events), nil, nil
	}
	events, cmds, err := Translate(raw)
	events = d.ownedResultEvents(raw, events)
	d.observeEvents(events)
	return events, cmds, err
}

func (d *protocolDriver) handleNotification(method codexMethod, raw agentbridge.RawEvent) ([]agentbridge.Event, bool) {
	if d.owned.id != "" {
		if events, handled := d.ownedMessageEvents(method, params(raw)); handled {
			d.observeEvents(events)
			return events, true
		}
	}
	if method == codexMethodError {
		p := params(raw)
		errText := codexNotificationErrorMessage(p)
		if retry, _ := p["willRetry"].(bool); retry {
			return []agentbridge.Event{{Kind: agentbridge.EventWarning, Err: errText}}, true
		}
		d.recordRuntimeError(errText)
		return []agentbridge.Event{{Kind: agentbridge.EventError, Err: errText}}, true
	}
	if method == codexMethodTurnStarted || method == codexMethodTurnStartedSlash {
		d.turnStarted = true
	}
	if method == codexMethodTurnCompleted || method == codexMethodTurnCompleteSlash {
		p := params(raw)
		if stringField(mapField(p, "turn"), "status") != "" {
			// Native terminal status is authoritative; legacy error fallback must
			// neither overwrite interruption nor reject a successful retry.
			return nil, false
		}
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
