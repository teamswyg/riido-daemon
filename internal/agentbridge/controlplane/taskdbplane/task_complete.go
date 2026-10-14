package taskdbplane

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func (p *Plane) CompleteTask(ctx context.Context, taskID string, res agentbridge.Result) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.withFileLock(ctx, func() error {
		now := p.now().UTC()
		db, err := taskdb.LoadTaskDB(p.path)
		if err != nil {
			return planeWrapf(ErrTaskDBPlanePersistence, "complete-task.load-task-db", err, "load task DB")
		}
		updated, err := applyTerminalResult(db, taskID, res, now)
		if err != nil {
			return err
		}
		leases, err := loadRuntimeLeaseRegistryOrEmpty(p.leasePath)
		if err != nil {
			return err
		}
		if taskDBChanged(db, updated, taskID) {
			report, _ := controlplane.TaskReportContextFromContext(ctx)
			if _, err := requireActiveRuntimeLease(leases, taskID, now, report); err != nil {
				return err
			}
		}
		if err := taskdb.SaveTaskDB(p.path, updated); err != nil {
			return planeWrapf(ErrTaskDBPlanePersistence, "complete-task.save-task-db", err, "save task DB")
		}
		return p.releaseCompletedLease(leases, taskID, now)
	})
}
