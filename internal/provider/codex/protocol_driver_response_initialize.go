package codex

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *protocolDriver) handleInitializeResponse(ctx context.Context, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	if err := d.sendNotification(ctx, io, codexMethodInitialized, map[string]any{}); err != nil {
		return nil, nil, err
	}
	d.initialized = true
	method, params := d.threadStartRequest()
	if _, err := d.sendRequest(ctx, io, method, params); err != nil {
		return nil, nil, err
	}
	return nil, nil, nil
}

func (d *protocolDriver) threadStartRequest() (codexMethod, map[string]any) {
	method := codexMethodThreadStart
	params := map[string]any{}
	if d.req.ResumeSessionID != "" {
		method = codexMethodThreadResume
		params["threadId"] = d.req.ResumeSessionID
	} else if d.req.Model != "" {
		params["model"] = d.req.Model
	}
	return method, params
}
