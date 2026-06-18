package saasplane

import (
	"context"
	"net/url"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func (p *Plane) pollAgent(ctx context.Context, agentID, runtimeID string, wait time.Duration) (assignmentcontract.PollResponse, error) {
	var out assignmentcontract.PollResponse
	err := p.postJSON(ctx, "/v1/agents/"+url.PathEscape(agentID)+"/poll", assignmentcontract.PollRequest{
		DaemonID:  p.cfg.DaemonID,
		DeviceID:  p.cfg.DeviceID,
		RuntimeID: runtimeID,
		WaitMs:    pollWaitMilliseconds(wait),
	}, &out)
	return out, err
}

func pollWaitMilliseconds(wait time.Duration) int {
	if wait <= 0 {
		return 0
	}
	milliseconds := wait.Milliseconds()
	if milliseconds <= 0 {
		return 1
	}
	return int(milliseconds)
}
