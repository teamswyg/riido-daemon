package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

type claimState struct {
	db     taskdb.TaskDB
	leases RuntimeLeaseRegistry
	now    time.Time
}

func (p *Plane) loadClaimState() (claimState, error) {
	db, err := taskdb.LoadTaskDBOrEmpty(p.path)
	if err != nil {
		return claimState{}, planeWrapf(ErrTaskDBPlanePersistence, "claim-task.load-task-db", err, "load task DB")
	}
	leases, err := loadRuntimeLeaseRegistryOrEmpty(p.leasePath)
	if err != nil {
		return claimState{}, err
	}
	return claimState{db: db, leases: leases, now: p.now().UTC()}, nil
}

func (p *Plane) reconcileClaimState(state claimState) (claimState, error) {
	db, leases, changed, err := reconcileExpiredRuntimeLeases(state.db, state.leases, state.now)
	if err != nil {
		return claimState{}, err
	}
	state.db = db
	state.leases = leases
	if !changed {
		return state, nil
	}
	if err := taskdb.SaveTaskDB(p.path, state.db); err != nil {
		return claimState{}, planeWrapf(ErrTaskDBPlanePersistence, "claim-task.save-task-db", err, "save task DB after lease reconciliation")
	}
	if err := saveRuntimeLeaseRegistry(p.leasePath, p.path, state.leases, state.now); err != nil {
		return claimState{}, err
	}
	return state, nil
}
