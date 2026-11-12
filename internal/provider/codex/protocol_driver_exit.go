package codex

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// OnProcessExit emits one Error event per still-pending request so
// callers never block forever waiting for a response that will never
// arrive. Also clears the pending map (cleanup).
func (d *protocolDriver) OnProcessExit(_ context.Context, status agentbridge.ProcessExitStatus, _ agentbridge.ProtocolIO) ([]agentbridge.Event, error) {
	if d.owned.id != "" && !d.owned.done {
		message := fmt.Sprintf("codex: process exited before owned turn completed (thread=%s turn=%s exit_code=%d)", d.threadID, d.owned.id, status.Code)
		if d.lastRuntimeError != "" {
			message += ": " + d.lastRuntimeError
		}
		return d.failedEvents(message), nil
	}
	if d.lastRuntimeError != "" && !d.sawAssistantOutput {
		return d.failedEvents(d.lastRuntimeError), nil
	}
	if len(d.pending) == 0 {
		return nil, nil
	}
	out := make([]agentbridge.Event, 0, len(d.pending))
	for id, pr := range d.pending {
		if pr.method == codexMethodThreadNameSet {
			continue
		}
		out = append(out, agentbridge.Event{
			Kind: agentbridge.EventError,
			Err:  fmt.Sprintf("codex: pending RPC request id=%d method=%s cancelled by process exit code=%d", id, pr.method, status.Code),
		})
	}
	d.pending = map[int64]pendingRequest{}
	return out, nil
}

// OnClose releases any final state. The pending map is already cleared
// by OnProcessExit; if the session terminates for a non-exit reason
// (Cancel/Timeout) we still want pending entries dropped so future
// re-use can't leak them.
func (d *protocolDriver) OnClose(_ context.Context, _ agentbridge.ProtocolIO) error {
	d.pending = nil
	return nil
}
