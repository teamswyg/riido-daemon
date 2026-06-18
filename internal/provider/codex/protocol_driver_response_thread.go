package codex

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *protocolDriver) handleThreadReadyResponse(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	d.threadID = threadIDFromResult(mapField(raw.Payload, "result"))
	events := d.threadIdentifiedEvents()
	if _, err := d.sendRequest(ctx, io, codexMethodTurnStart, d.turnStartParams()); err != nil {
		return nil, nil, err
	}
	return events, nil, nil
}

func (d *protocolDriver) threadIdentifiedEvents() []agentbridge.Event {
	if d.threadID == "" {
		return nil
	}
	return []agentbridge.Event{{
		Kind:      agentbridge.EventSessionIdentified,
		SessionID: d.threadID,
	}}
}

func (d *protocolDriver) turnStartParams() map[string]any {
	params := map[string]any{}
	if d.threadID != "" {
		params["threadId"] = d.threadID
	}
	if d.req.Prompt != "" {
		params["input"] = []map[string]any{{"type": "text", "text": d.req.Prompt}}
	}
	if d.req.Model != "" {
		params["model"] = d.req.Model
	}
	return params
}

func (d *protocolDriver) handleTurnStartResponse() ([]agentbridge.Event, []agentbridge.Command, error) {
	d.turnStarted = true
	return []agentbridge.Event{{
		Kind:  agentbridge.EventLifecycle,
		Phase: agentbridge.StateRunning,
	}}, nil, nil
}
