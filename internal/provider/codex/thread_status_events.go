package codex

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// threadStatusEvents maps Codex thread/status/changed notifications to
// run-scope events. Completion is gated on turnStarted so initial idle cannot
// end the run before any work happens.
func (d *protocolDriver) threadStatusEvents(p map[string]any) []agentbridge.Event {
	status := threadStatus(strings.ToLower(strings.TrimSpace(threadStatusFromPayload(p))))
	switch {
	case codexStatusIsError(status):
		return d.failedEvents("codex thread status: " + string(status))
	case codexStatusIsTerminal(status):
		if !d.turnStarted {
			return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex thread status: " + string(status) + " (no active turn)"}}
		}
		d.turnStarted = false
		return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}}}
	case codexStatusIsActive(status):
		d.turnStarted = true
		return []agentbridge.Event{{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}}
	default:
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex thread status changed: " + string(status)}}
	}
}

func threadStatusFromPayload(p map[string]any) string {
	if s := stringField(p, "status"); s != "" {
		return s
	}
	if s := stringField(mapField(p, "status"), "type"); s != "" {
		return s
	}
	if s := stringField(mapField(p, "thread"), "status"); s != "" {
		return s
	}
	return stringField(p, "state")
}
