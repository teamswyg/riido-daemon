package bridge

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
)

// Session is the caller-facing handle for one run.
type Session struct {
	inner       *session.Session
	droppedArgs []string
}

func newSession(inner *session.Session, droppedArgs []string) *Session {
	return &Session{inner: inner, droppedArgs: droppedArgs}
}

// Events returns the run-scope event stream, closed when the session ends.
func (s *Session) Events() <-chan agentbridge.Event { return s.inner.Events() }

// Result returns the single-value terminal result channel.
func (s *Session) Result() <-chan agentbridge.Result { return s.inner.Result() }

// Cancel signals the session to terminate as ResultCancelled.
func (s *Session) Cancel(cause error) { s.inner.Cancel(cause) }

// DroppedArgs returns the custom args that BuildStart removed because
// they collided with the adapter's BlockedArgs.
func (s *Session) DroppedArgs() []string { return s.droppedArgs }
