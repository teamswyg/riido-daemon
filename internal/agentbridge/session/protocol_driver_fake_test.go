package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// fakeDriver is the test scaffolding ProtocolDriver. Hooks are configurable
// and invocation counters are read only after the session actor terminates.
type fakeDriver struct {
	startStdin []byte

	onStart       func(ctx context.Context, io ProtocolIO) error
	onRaw         func(ctx context.Context, raw agentbridge.RawEvent, io ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error)
	onProcessExit func(ctx context.Context, status agentbridge.ProcessExitStatus, io ProtocolIO) ([]agentbridge.Event, error)
	onClose       func(ctx context.Context, io ProtocolIO) error

	startCalls int
	rawCalls   int
	exitCalls  int
	closeCalls int
}

func (d *fakeDriver) OnStart(ctx context.Context, io ProtocolIO) error {
	d.startCalls++
	if d.onStart != nil {
		return d.onStart(ctx, io)
	}
	if len(d.startStdin) > 0 {
		return io.WriteStdin(ctx, d.startStdin)
	}
	return nil
}

func (d *fakeDriver) OnRaw(
	ctx context.Context,
	raw agentbridge.RawEvent,
	io ProtocolIO,
) ([]agentbridge.Event, []agentbridge.Command, error) {
	d.rawCalls++
	if d.onRaw != nil {
		return d.onRaw(ctx, raw, io)
	}
	return nil, nil, nil
}

func (d *fakeDriver) OnProcessExit(
	ctx context.Context,
	status agentbridge.ProcessExitStatus,
	io ProtocolIO,
) ([]agentbridge.Event, error) {
	d.exitCalls++
	if d.onProcessExit != nil {
		return d.onProcessExit(ctx, status, io)
	}
	return nil, nil
}

func (d *fakeDriver) OnClose(ctx context.Context, io ProtocolIO) error {
	d.closeCalls++
	if d.onClose != nil {
		return d.onClose(ctx, io)
	}
	return nil
}
