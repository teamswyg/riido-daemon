package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// Session is the public handle the caller uses to consume events and result.
type Session struct {
	events chan agentbridge.Event
	result chan agentbridge.Result
	cancel chan cancelRequest
	done   chan struct{}

	running process.RunningProcess
}

type cancelRequest struct {
	ctx   context.Context
	cause error
}

// Events returns the event stream. It is closed when the session terminates.
func (s *Session) Events() <-chan agentbridge.Event { return s.events }

// Result returns one terminal Result then closes.
func (s *Session) Result() <-chan agentbridge.Result { return s.result }

// Done closes when the session terminates without consuming Result.
func (s *Session) Done() <-chan struct{} { return s.done }

// Cancel signals the session to terminate as ResultCancelled.
func (s *Session) Cancel(cause error) {
	s.CancelWithContext(context.Background(), cause)
}

// CancelWithContext preserves lifecycle authority such as forced shutdown.
func (s *Session) CancelWithContext(ctx context.Context, cause error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if lifecycle.FromContext(ctx).ShutdownLevel().IsForced() && s.running != nil {
		go func() { _ = s.running.Kill(ctx) }()
	}
	select {
	case s.cancel <- cancelRequest{ctx: ctx, cause: cause}:
	default:
	}
}

// runningForTest exposes the underlying process for whitebox tests.
func (s *Session) runningForTest() *process.FakeRunning {
	r, _ := s.running.(*process.FakeRunning)
	return r
}
