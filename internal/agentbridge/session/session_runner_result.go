package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (r *sessionRunner) finalResult() agentbridge.Result {
	finalResult := r.state.Result
	finalResult.SessionID = r.state.SessionID
	finalResult.Usage = r.state.Usage
	if finalResult.Workdir == "" {
		finalResult.Workdir = r.cfg.Spawn.Dir
	}
	if finalResult.StartedAt.IsZero() {
		finalResult.StartedAt = r.startedAt
	}
	if finalResult.FinishedAt.IsZero() {
		finalResult.FinishedAt = r.cfg.Now()
	}
	return finalResult
}
