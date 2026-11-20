package session

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Validate before publishing: a foreign pin must never enter the server's
// assignment history, including for providers without a protocol driver.
func (r *sessionRunner) guardSessionIdentity(ev agentbridge.Event) (agentbridge.Event, bool) {
	if ev.NativeHistory != nil && (r.state.SessionID == "" || ev.NativeHistory.SessionID != r.state.SessionID || ev.NativeHistory.SessionID != r.cfg.Request.ResumeSessionID) {
		r.sessionIdentityBlocked = true
		return agentbridge.Event{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultBlocked, Error: "native history session identity mismatch"}}, true
	}
	if ev.Kind == agentbridge.EventResult && (ev.Result.Status == agentbridge.ResultCompleted || ev.Result.Status == "") &&
		strings.TrimSpace(r.cfg.Request.ResumeSessionID) != "" && r.state.SessionID == "" {
		r.state.Output = nil
		r.sessionIdentityBlocked = true
		return agentbridge.Event{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultBlocked, Error: "provider did not confirm the requested session; history continuity is unverified"}}, true
	}
	if ev.Kind != agentbridge.EventSessionIdentified {
		return ev, true
	}
	id := strings.TrimSpace(ev.SessionID)
	if id == "" || id == r.state.SessionID {
		return ev, false
	}
	expected := r.state.SessionID
	if expected == "" {
		expected = strings.TrimSpace(r.cfg.Request.ResumeSessionID)
	}
	if expected != "" && expected != id {
		r.state.Output = nil
		r.sessionIdentityBlocked = true
		return agentbridge.Event{Kind: agentbridge.EventResult, Result: agentbridge.Result{
			Status: agentbridge.ResultBlocked,
			Error:  "provider session identity changed; refusing to replace task history",
		}}, true
	}
	ev.SessionID = id
	return ev, true
}
