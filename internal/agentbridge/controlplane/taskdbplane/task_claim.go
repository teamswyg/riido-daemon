package taskdbplane

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func (p *Plane) ClaimTask(ctx context.Context, runtimeID string) (*bridge.TaskRequest, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if runtimeID == "" {
		return nil, planeErrorf(ErrTaskDBPlaneRuntime, "claim-task", "empty RuntimeID")
	}
	var req *bridge.TaskRequest
	err := p.withFileLock(ctx, func() error {
		var err error
		req, err = p.claimTaskLocked(runtimeID)
		return err
	})
	return req, err
}
