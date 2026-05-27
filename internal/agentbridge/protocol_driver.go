package agentbridge

import "context"

// ProcessExitStatus is the provider-neutral process-exit signal passed to a
// ProtocolDriver. The concrete process port owns OS/process details; providers
// only need the normalized exit code and optional diagnostic string.
type ProcessExitStatus struct {
	Code int
	Err  string
}

// ProtocolDriver is the run-level transport hook a provider adapter can
// install when its protocol needs an active handshake, such as Codex app-server
// JSON-RPC or Claude's stdin frame followed by EOF.
//
// This is a provider-runtime port, not a session actor implementation detail.
// Drivers may keep provider transport bookkeeping, but they must return Events
// and Commands through the normal reducer path rather than mutating RunState.
type ProtocolDriver interface {
	OnStart(ctx context.Context, io ProtocolIO) error
	OnRaw(ctx context.Context, raw RawEvent, io ProtocolIO) ([]Event, []Command, error)
	OnProcessExit(ctx context.Context, status ProcessExitStatus, io ProtocolIO) ([]Event, error)
	OnClose(ctx context.Context, io ProtocolIO) error
}

// ProtocolIO is the only sanctioned side-effect surface a ProtocolDriver gets:
// writing bytes to the already-spawned provider process and closing stdin.
type ProtocolIO interface {
	WriteStdin(ctx context.Context, b []byte) error
	CloseStdin(ctx context.Context) error
}

// ProtocolDriverProvider is the optional adapter interface used by runtime
// orchestration. A provider that needs an active transport handshake implements
// this in addition to Adapter.
type ProtocolDriverProvider interface {
	NewProtocolDriver(req StartRequest) (ProtocolDriver, error)
}
