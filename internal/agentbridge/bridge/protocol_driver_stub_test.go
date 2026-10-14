package bridge

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type protocolAdapter struct {
	stubAdapter
	driver agentbridge.ProtocolDriver
}

func (a *protocolAdapter) NewProtocolDriver(
	_ agentbridge.StartRequest,
) (agentbridge.ProtocolDriver, error) {
	return a.driver, nil
}

type driverSpy struct {
	started chan struct{}
}

func (d *driverSpy) OnStart(context.Context, agentbridge.ProtocolIO) error {
	close(d.started)
	return nil
}

func (d *driverSpy) OnRaw(
	_ context.Context,
	raw agentbridge.RawEvent,
	_ agentbridge.ProtocolIO,
) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "chunk" {
		return []agentbridge.Event{{
			Kind: agentbridge.EventResult,
			Result: agentbridge.Result{
				Status: agentbridge.ResultCompleted,
				Output: string(raw.Bytes),
			},
		}}, nil, nil
	}
	return nil, nil, nil
}

func (d *driverSpy) OnProcessExit(
	context.Context,
	agentbridge.ProcessExitStatus,
	agentbridge.ProtocolIO,
) ([]agentbridge.Event, error) {
	return nil, nil
}

func (d *driverSpy) OnClose(context.Context, agentbridge.ProtocolIO) error { return nil }
