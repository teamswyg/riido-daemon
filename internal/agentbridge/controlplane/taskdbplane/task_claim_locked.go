package taskdbplane

import "github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"

func (p *Plane) claimTaskLocked(runtimeID string) (*bridge.TaskRequest, error) {
	if err := p.reloadRuntimeRegistry(); err != nil {
		return nil, err
	}
	state, err := p.loadClaimState()
	if err != nil {
		return nil, err
	}
	state, err = p.reconcileClaimState(state)
	if err != nil {
		return nil, err
	}
	for _, record := range claimCandidates(state.db) {
		req, claimed, err := p.tryClaimRecord(state.db, &state.leases, runtimeID, record)
		if err != nil {
			return nil, err
		}
		if claimed {
			return req, nil
		}
	}
	return nil, nil
}
