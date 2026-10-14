package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func (p *Plane) tryClaimRecord(db taskdb.TaskDB, leases *RuntimeLeaseRegistry, runtimeID string, record taskdb.TaskRecord) (*bridge.TaskRequest, bool, error) {
	provider, prompt, ok := claimRecordInputs(db, record)
	if !ok {
		return nil, false, nil
	}
	selection, ok := p.runtimeSelectionForTask(provider, runtimeID)
	if !ok {
		return nil, false, nil
	}
	approvalID := approvalIDForTask(db, record.ID)
	if requiresApproval(db, record) && approvalID == "" {
		return nil, false, nil
	}
	now := p.now().UTC()
	updated, ok := applyClaimTransition(db, record, runtimeID, approvalID, now)
	if !ok {
		return nil, false, nil
	}
	var lease RuntimeLeaseRecord
	var leased bool
	*leases, lease, leased = acquireRuntimeLease(*leases, record.ID, runtimeID, string(selection.Runtime.CapabilityFingerprint), now, p.leaseTTL)
	if !leased {
		return nil, false, nil
	}
	if err := saveRuntimeLeaseRegistry(p.leasePath, p.path, *leases, now); err != nil {
		return nil, false, err
	}
	if err := taskdb.SaveTaskDB(p.path, updated); err != nil {
		return nil, false, planeWrapf(ErrTaskDBPlanePersistence, "claim-task.save-task-db", err, "save claimed task DB")
	}
	req := taskRequestFromRecord(p.path, record, provider, prompt, lease)
	return &req, true, nil
}

func applyClaimTransition(db taskdb.TaskDB, record taskdb.TaskRecord, runtimeID, approvalID string, now time.Time) (taskdb.TaskDB, bool) {
	updated, _, _, err := taskdb.ApplyGuardedTaskTransition(db, taskdb.TaskTransitionInput{
		TaskID:  record.ID,
		ToState: task.StateClaimed,
		Event:   ir.EventTaskClaimed,
		Actor:   defaultActor,
		Source:  sourceName,
		Reason:  defaultClaimReason + ": " + runtimeID,
		Guard:   guardFor(db, record, "claim:"+runtimeID, approvalID),
	}, now)
	return updated, err == nil
}

func claimRecordInputs(db taskdb.TaskDB, record taskdb.TaskRecord) (provider, prompt string, ok bool) {
	provider = providerFor(db, record)
	if provider == "" || !providerAvailable(db, provider) {
		return "", "", false
	}
	prompt = promptFor(record)
	if prompt == "" {
		return "", "", false
	}
	return provider, prompt, true
}
