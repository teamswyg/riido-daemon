package taskdbplane

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func (p *Plane) withDB(ctx context.Context, taskID string, mutator func(taskdb.TaskDB, time.Time) (taskdb.TaskDB, error)) error {
	return p.withFileLock(ctx, func() error {
		now := p.now().UTC()
		db, err := taskdb.LoadTaskDB(p.path)
		if err != nil {
			return planeWrapf(ErrTaskDBPlanePersistence, "with-db.load-task-db", err, "load task DB")
		}
		updated, err := mutator(db, now)
		if err != nil {
			return err
		}
		if taskDBChanged(db, updated, taskID) {
			if err := p.requireLeaseForChangedTask(ctx, taskID, now); err != nil {
				return err
			}
		}
		if err := taskdb.SaveTaskDB(p.path, updated); err != nil {
			return planeWrapf(ErrTaskDBPlanePersistence, "with-db.save-task-db", err, "save task DB")
		}
		return nil
	})
}

func (p *Plane) requireLeaseForChangedTask(ctx context.Context, taskID string, now time.Time) error {
	leases, err := loadRuntimeLeaseRegistryOrEmpty(p.leasePath)
	if err != nil {
		return err
	}
	report, _ := controlplane.TaskReportContextFromContext(ctx)
	_, err = requireActiveRuntimeLease(leases, taskID, now, report)
	return err
}
