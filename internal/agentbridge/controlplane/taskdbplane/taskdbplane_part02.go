package taskdbplane

import (
	"context"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
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

func (p *Plane) claimTaskLocked(runtimeID string) (*bridge.TaskRequest, error) {
	if err := p.reloadRuntimeRegistry(); err != nil {
		return nil, err
	}
	db, err := taskdb.LoadTaskDBOrEmpty(p.path)
	if err != nil {
		return nil, planeWrapf(ErrTaskDBPlanePersistence, "claim-task.load-task-db", err, "load task DB")
	}
	leases, err := loadRuntimeLeaseRegistryOrEmpty(p.leasePath)
	if err != nil {
		return nil, err
	}
	now := p.now().UTC()
	db, leases, changed, err := reconcileExpiredRuntimeLeases(db, leases, now)
	if err != nil {
		return nil, err
	}
	if changed {
		if err := taskdb.SaveTaskDB(p.path, db); err != nil {
			return nil, planeWrapf(ErrTaskDBPlanePersistence, "claim-task.save-task-db", err, "save task DB after lease reconciliation")
		}
		if err := saveRuntimeLeaseRegistry(p.leasePath, p.path, leases, now); err != nil {
			return nil, err
		}
	}
	candidates := claimCandidates(db)
	for _, record := range candidates {
		provider := providerFor(db, record)
		if provider == "" || !providerAvailable(db, provider) {
			continue
		}
		selection, ok := p.runtimeSelectionForTask(provider, runtimeID)
		if !ok {
			continue
		}
		prompt := promptFor(record)
		if prompt == "" {
			continue
		}
		approvalID := approvalIDForTask(db, record.ID)
		if requiresApproval(db, record) && approvalID == "" {
			continue
		}
		now := p.now().UTC()
		input := taskdb.TaskTransitionInput{
			TaskID:  record.ID,
			ToState: task.StateClaimed,
			Event:   ir.EventTaskClaimed,
			Actor:   defaultActor,
			Source:  sourceName,
			Reason:  defaultClaimReason + ": " + runtimeID,
			Guard:   guardFor(db, record, "claim:"+runtimeID, approvalID),
		}
		updated, _, _, err := taskdb.ApplyGuardedTaskTransition(db, input, now)
		if err != nil {
			continue
		}
		var lease RuntimeLeaseRecord
		leases, lease, ok = acquireRuntimeLease(leases, record.ID, runtimeID, string(selection.Runtime.CapabilityFingerprint), now, p.leaseTTL)
		if !ok {
			continue
		}
		if err := saveRuntimeLeaseRegistry(p.leasePath, p.path, leases, now); err != nil {
			return nil, err
		}
		if err := taskdb.SaveTaskDB(p.path, updated); err != nil {
			return nil, planeWrapf(ErrTaskDBPlanePersistence, "claim-task.save-task-db", err, "save claimed task DB")
		}
		req := taskRequestFromRecord(p.path, record, provider, prompt, lease)
		return &req, nil
	}
	return nil, nil
}

func (p *Plane) WatchCancellation(_ context.Context, _ string) (<-chan error, error) {
	ch := make(chan error)
	close(ch)
	return ch, nil
}

func (p *Plane) StartTask(ctx context.Context, taskID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return p.withDB(ctx, taskID, func(db taskdb.TaskDB, now time.Time) (taskdb.TaskDB, error) {
		return ensurePreparing(db, taskID, now)
	})
}

func (p *Plane) ReportEvent(ctx context.Context, taskID string, ev agentbridge.Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if ev.Kind != agentbridge.EventLifecycle || ev.Phase != agentbridge.StateRunning {
		return nil
	}
	return p.withDB(ctx, taskID, func(db taskdb.TaskDB, now time.Time) (taskdb.TaskDB, error) {
		return ensureRunning(db, taskID, now)
	})
}

func (p *Plane) CompleteTask(ctx context.Context, taskID string, res agentbridge.Result) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
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
		mutated := taskDBChanged(db, updated, taskID)
		leases, err := loadRuntimeLeaseRegistryOrEmpty(p.leasePath)
		if err != nil {
			return err
		}
		if mutated {
			report, _ := controlplane.TaskReportContextFromContext(ctx)
			if _, err := requireActiveRuntimeLease(leases, taskID, now, report); err != nil {
				return err
			}
		}
		if err := taskdb.SaveTaskDB(p.path, updated); err != nil {
			return planeWrapf(ErrTaskDBPlanePersistence, "complete-task.save-task-db", err, "save task DB")
		}
		leases, changed := releaseRuntimeLease(leases, taskID, now)
		if !changed {
			return nil
		}
		return saveRuntimeLeaseRegistry(p.leasePath, p.path, leases, now)
	})
}
