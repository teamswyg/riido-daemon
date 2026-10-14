package runtimeactor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
)

// SessionHandle is the caller-facing per-task handle. Mirrors
// session.Session but is the Actor's surface so we don't leak the
// internal session package across the API boundary.
type SessionHandle struct {
	TaskID  string
	session *session.Session
}

// Events returns the run-scope event stream, closed when the session
// terminates.
func (h *SessionHandle) Events() <-chan agentbridge.Event { return h.session.Events() }

// Result returns the terminal result channel (single value, then closed).
func (h *SessionHandle) Result() <-chan agentbridge.Result { return h.session.Result() }

// Done signals termination without consuming Result. Used by the Actor
// itself; callers normally prefer Result().
func (h *SessionHandle) Done() <-chan struct{} { return h.session.Done() }
