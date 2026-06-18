package claude

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// protocolDriver implements agentbridge.ProtocolDriver for Claude Code's
// `claude -p --input-format stream-json` mode.
type protocolDriver struct {
	req     agentbridge.StartRequest
	written bool
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
