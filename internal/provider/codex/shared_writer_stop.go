package codex

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *protocolDriver) OnTerminate(ctx context.Context, io agentbridge.ProtocolIO) (bool, error) {
	s := d.shared
	if s == nil || s.done || s.stopping {
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
