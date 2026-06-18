package taskdbplane

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func (p *Plane) StartTask(ctx context.Context, taskID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.withDB(ctx, taskID, func(db taskdb.TaskDB, now time.Time) (taskdb.TaskDB, error) {
		return ensurePreparing(db, taskID, now)
	})
}

func (p *Plane) ReportEvent(ctx context.Context, taskID string, ev agentbridge.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if ev.Kind != agentbridge.EventLifecycle || ev.Phase != agentbridge.StateRunning {
		return nil
	}
	return p.withDB(ctx, taskID, func(db taskdb.TaskDB, now time.Time) (taskdb.TaskDB, error) {
		return ensureRunning(db, taskID, now)
	})
}
