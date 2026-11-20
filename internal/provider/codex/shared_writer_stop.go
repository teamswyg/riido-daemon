package codex

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *protocolDriver) OnTerminate(ctx context.Context, io agentbridge.ProtocolIO) (bool, error) {
	if err := d.cancelMCPApprovals(ctx, io); err != nil {
		return false, err
	}
	s := d.shared
	if s == nil || s.done {
		return false, nil
	}
	if s.stopping {
		// Failure cleanup may already be in flight. Cancellation must await it
		// and must never turn its acknowledgement into a reconnect attempt.
		s.startError = ""
		for _, pending := range d.pending {
			if pending.method == codexMethodQueueDelete {
				return true, nil
			}
		}
		return false, nil
	}
	s.stopping = true
	if s.submissionID != "" {
		return true, d.deleteSharedQueue(ctx, io)
	}
	for _, pending := range d.pending {
		if pending.method == codexMethodQueueAdd {
			return true, nil
		}
	}
	if s.submitted {
		return false, errors.New(sharedStopNotice)
	}
	return false, nil
}

func (d *protocolDriver) deleteSharedQueue(ctx context.Context, io agentbridge.ProtocolIO) error {
	_, err := d.sendRequest(ctx, io, codexMethodQueueDelete, map[string]any{"threadId": d.threadID, "queuedSubmissionId": d.shared.submissionID})
	return err
}
