package taskdbplane

import (
	"context"
	"sort"
	"strconv"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/supervisor"
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
			leases, err := loadRuntimeLeaseRegistryOrEmpty(p.leasePath)
			if err != nil {
				return err
			}
			report, _ := controlplane.TaskReportContextFromContext(ctx)
			if _, err := requireActiveRuntimeLease(leases, taskID, now, report); err != nil {
				return err
			}
		}
		if err := taskdb.SaveTaskDB(p.path, updated); err != nil {
			return planeWrapf(ErrTaskDBPlanePersistence, "with-db.save-task-db", err, "save task DB")
		}
		return nil
	})
}

func claimCandidates(db taskdb.TaskDB) []taskdb.TaskRecord {
	out := make([]taskdb.TaskRecord, 0, len(db.Tasks))
	for _, record := range db.Tasks {
		if record.State.Code() == task.TaskStateCodeQueued {
			out = append(out, record)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		left := out[i].UpdatedAt
		right := out[j].UpdatedAt
		if left != right {
			return left < right
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func taskRequestFromRecord(path string, record taskdb.TaskRecord, provider, prompt string, lease RuntimeLeaseRecord) bridge.TaskRequest {
	meta := map[string]string{
		metadataTaskDB: path,
	}
	if record.ProjectID != "" {
		meta[supervisor.MetadataWorkspaceID] = record.ProjectID
	}
	if record.SourceDocumentPath != "" {
		meta[metadataDocument] = record.SourceDocumentPath
	}
	if lease.LeaseID != "" {
		meta[controlplane.MetadataRuntimeLeaseID] = lease.LeaseID
		meta[controlplane.MetadataRuntimeFencingToken] = strconv.FormatInt(lease.FencingToken, 10)
		if lease.CapabilityFingerprint != "" {
			meta[controlplane.MetadataRuntimeCapabilityFingerprint] = lease.CapabilityFingerprint
		}
	}
	return bridge.TaskRequest{
		ID:       record.ID,
		Provider: bridge.Provider(provider),
		Prompt:   prompt,
		Metadata: meta,
	}
}

func ensurePreparing(db taskdb.TaskDB, taskID string, now time.Time) (taskdb.TaskDB, error) {
	record, ok := findTask(db, taskID)
	if !ok {
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "ensure-preparing", "task %s not found", taskID)
	}
	switch record.State.Code() {
	case task.TaskStateCodePreparing, task.TaskStateCodeRunning, task.TaskStateCodeNeedsInput, task.TaskStateCodeBlocked, task.TaskStateCodeValidating, task.TaskStateCodePatchReady, task.TaskStateCodeHumanReview:
		return db, nil
	case task.TaskStateCodeClaimed:
		return applyTransition(db, record, task.StatePreparing, ir.EventWorkdirPreparing, "workspace preparation started", "preparing", now)
	default:
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "ensure-preparing", "cannot start task %s from state %s", taskID, record.State)
	}
}

func ensureRunning(db taskdb.TaskDB, taskID string, now time.Time) (taskdb.TaskDB, error) {
	record, ok := findTask(db, taskID)
	if !ok {
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "ensure-running", "task %s not found", taskID)
	}
	switch record.State.Code() {
	case task.TaskStateCodeRunning, task.TaskStateCodeNeedsInput, task.TaskStateCodeBlocked, task.TaskStateCodeValidating, task.TaskStateCodePatchReady, task.TaskStateCodeHumanReview:
		return db, nil
	case task.TaskStateCodeClaimed:
		var err error
		db, err = applyTransition(db, record, task.StatePreparing, ir.EventWorkdirPreparing, "workspace preparation started", "preparing", now)
		if err != nil {
			return taskdb.TaskDB{}, err
		}
		record, _ = findTask(db, taskID)
		fallthrough
	case task.TaskStateCodePreparing:
		return applyTransition(db, record, task.StateRunning, ir.EventRunStarted, "provider process started", "run-started", now)
	default:
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "ensure-running", "cannot run task %s from state %s", taskID, record.State)
	}
}

func applyTerminalResult(db taskdb.TaskDB, taskID string, res agentbridge.Result, now time.Time) (taskdb.TaskDB, error) {
	record, ok := findTask(db, taskID)
	if !ok {
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "apply-terminal-result", "task %s not found", taskID)
	}
	if record.State.IsTerminal() {
		return db, nil
	}
	recordState := record.State.Code()
	status := res.Status
	if status == "" {
		status = agentbridge.ResultCompleted
	}
	switch status {
	case agentbridge.ResultCompleted:
		if recordState == task.TaskStateCodeValidating || recordState == task.TaskStateCodePatchReady || recordState == task.TaskStateCodeHumanReview {
			return db, nil
		}
		var err error
		db, err = ensureRunning(db, taskID, now)
		if err != nil {
			return taskdb.TaskDB{}, err
		}
		record, _ = findTask(db, taskID)
		return applyTransition(db, record, task.StateValidating, ir.EventRunReportedDone, "provider reported run done", "run-reported-done", now)
	case agentbridge.ResultCancelled:
		return applyTransition(db, record, task.StateCancelled, ir.EventTaskCancelled, resultReason(res, "provider run cancelled"), "cancelled", now)
	case agentbridge.ResultTimeout:
		if timeoutCanOriginate(record.State) {
			return applyTransition(db, record, task.StateTimedOut, ir.EventTaskTimedOut, resultReason(res, "provider run timed out"), "timed-out", now)
		}
		return applyTransition(db, record, task.StateFailed, ir.EventTaskFailed, resultReason(res, "provider timed out before running"), "failed-timeout-before-running", now)
	case agentbridge.ResultBlocked:
		if recordState == task.TaskStateCodeBlocked {
			return db, nil
		}
		if recordState == task.TaskStateCodeClaimed {
			var err error
			db, err = ensurePreparing(db, taskID, now)
			if err != nil {
				return taskdb.TaskDB{}, err
			}
			record, _ = findTask(db, taskID)
			recordState = record.State.Code()
		}
		if recordState == task.TaskStateCodePreparing || recordState == task.TaskStateCodeRunning {
			return applyTransition(db, record, task.StateBlocked, ir.EventBlockerRaised, resultReason(res, "runtime eligibility blocked task"), "blocked", now)
		}
		return applyTransition(db, record, task.StateFailed, ir.EventTaskFailed, resultReason(res, "runtime eligibility blocked task from invalid state"), "failed:blocked-invalid-state", now)
	default:
		return applyTransition(db, record, task.StateFailed, ir.EventTaskFailed, resultReason(res, string(status)), "failed:"+string(status), now)
	}
}
