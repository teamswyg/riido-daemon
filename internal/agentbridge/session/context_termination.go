package session

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

var errRunHardDeadline = errors.New("run hard timeout exceeded")

func contextTerminationEvent(ctx context.Context) agentbridge.Event {
	cause := context.Cause(ctx)
	if cause == nil {
		cause = context.Canceled
	}
	kind := agentbridge.EventCancellation
	if errors.Is(cause, errRunHardDeadline) {
		kind = agentbridge.EventTimeout
	}
	return agentbridge.Event{Kind: kind, Err: cause.Error()}
}
