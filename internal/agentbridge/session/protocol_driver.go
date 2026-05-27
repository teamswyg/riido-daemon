package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// Compatibility aliases keep package-local tests and older internal callers on
// the same type names while the owning port lives in internal/agentbridge.
type ProtocolDriver = agentbridge.ProtocolDriver
type ProtocolIO = agentbridge.ProtocolIO
type ProtocolDriverProvider = agentbridge.ProtocolDriverProvider

// protocolIOImpl is the session's internal ProtocolIO bound to one
// process's stdin. The session actor is the sole owner — drivers
// receive it through their OnStart/OnRaw/... calls and never store it
// across goroutine boundaries.
type protocolIOImpl struct {
	proc process.RunningProcess
}

func newProtocolIO(proc process.RunningProcess) *protocolIOImpl {
	return &protocolIOImpl{proc: proc}
}

func (p *protocolIOImpl) WriteStdin(_ context.Context, b []byte) error {
	if p == nil || p.proc == nil {
		return errProtocolIONil
	}
	return p.proc.WriteStdin(b)
}

func (p *protocolIOImpl) CloseStdin(_ context.Context) error {
	if p == nil || p.proc == nil {
		return errProtocolIONil
	}
	return p.proc.CloseStdin()
}

// errProtocolIONil is a stable error value so drivers can match on it.
var errProtocolIONil = &protocolIOError{msg: "session: protocol io has no process"}

type protocolIOError struct{ msg string }

func (e *protocolIOError) Error() string { return e.msg }
