// Package taskdbplane adapts riido-task-db.v1 into the agentbridge
// control-plane ports.
//
// It is intentionally outside the core controlplane package: the
// port definitions stay independent from project persistence, while
// this adapter is allowed to translate taskdb.TaskRecord rows into
// bridge.TaskRequest values and report guarded TaskState transitions.
package taskdbplane

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/supervisor"
	"github.com/teamswyg/riido-daemon/internal/scheduling"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

const (
	RuntimeRegistrySchemaVersion      = "riido-runtime-registry.v1"
	RuntimeLeaseRegistrySchemaVersion = "riido-runtime-lease-registry.v1"

	sourceName         = "riido.agentbridge.taskdb"
	metadataTaskDB     = "task_db_path"
	metadataDocument   = "source_document_path"
	commandIDPrefix    = "command:riido.agentbridge.taskdb:"
	defaultActor       = "daemon"
	defaultClaimReason = "runtime claimed queued task DB row"

	defaultRuntimeLeaseTTL = 30 * time.Second
)

// RuntimeRegistry is the task DB source sidecar written next to the
// riido-task-db.v1 file. It lets local GUI/Zed integrations inspect
// runtime registration and heartbeat without reaching into daemon memory.
type RuntimeRegistry struct {
	SchemaVersion string                           `json:"schema_version"`
	TaskDBPath    string                           `json:"task_db_path"`
	UpdatedAt     time.Time                        `json:"updated_at"`
	Runtimes      []controlplane.RegisteredRuntime `json:"runtimes"`
}

// RuntimeLeaseRegistry is the task DB source sidecar that records the
// latest local C9 fencing token per task.
type RuntimeLeaseRegistry struct {
	SchemaVersion string               `json:"schema_version"`
	TaskDBPath    string               `json:"task_db_path"`
	UpdatedAt     time.Time            `json:"updated_at"`
	Leases        []RuntimeLeaseRecord `json:"leases"`
}

type RuntimeLeaseRecord struct {
	LeaseID               string     `json:"lease_id"`
	TaskID                string     `json:"task_id"`
	RuntimeID             string     `json:"runtime_id"`
	CapabilityFingerprint string     `json:"capability_fingerprint,omitempty"`
	ClaimedAt             time.Time  `json:"claimed_at"`
	LeaseUntil            time.Time  `json:"lease_until"`
	FencingToken          int64      `json:"fencing_token"`
	ReleasedAt            *time.Time `json:"released_at,omitempty"`
}

// Plane implements both TaskSourcePort and TaskReporterPort over one
// riido-task-db.v1 JSON file. The supervisor actor owns this value and
// calls it serially; the adapter therefore uses no mutex.
type Plane struct {
	path         string
	registryPath string
	leasePath    string
	lockPath     string
	leaseTTL     time.Duration
	now          func() time.Time
	runtimes     map[string]controlplane.RegisteredRuntime
}

func New(path string) (*Plane, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, planeErrorf(ErrTaskDBPlaneInput, "new", "empty task DB path")
	}
	if _, err := taskdb.LoadTaskDBOrEmpty(path); err != nil {
		return nil, planeWrapf(ErrTaskDBPlanePersistence, "new.load-task-db", err, "load task DB")
	}
	registryPath := runtimeRegistryPath(path)
	leasePath := runtimeLeaseRegistryPath(path)
	runtimes, err := loadRuntimeRegistryOrEmpty(registryPath)
	if err != nil {
		return nil, err
	}
	return &Plane{
		path:         path,
		registryPath: registryPath,
		leasePath:    leasePath,
		lockPath:     path + ".lock",
		leaseTTL:     defaultRuntimeLeaseTTL,
		now:          time.Now,
		runtimes:     runtimes,
	}, nil
}

func (p *Plane) RegisterRuntime(ctx context.Context, rt controlplane.RuntimeRegistration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if rt.RuntimeID == "" {
		return planeErrorf(ErrTaskDBPlaneRuntime, "register-runtime", "empty RuntimeID")
	}
	return p.withFileLock(ctx, func() error {
		if err := p.reloadRuntimeRegistry(); err != nil {
			return err
		}
		p.runtimes[rt.RuntimeID] = controlplane.RegisteredRuntime{
			RuntimeRegistration: rt,
			LastHeartbeat:       p.now().UTC(),
		}
		return p.saveRuntimeRegistry()
	})
}

func (p *Plane) DeregisterRuntime(ctx context.Context, runtimeID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if runtimeID == "" {
		return planeErrorf(ErrTaskDBPlaneRuntime, "deregister-runtime", "empty RuntimeID")
	}
	return p.withFileLock(ctx, func() error {
		if err := p.reloadRuntimeRegistry(); err != nil {
			return err
		}
		delete(p.runtimes, runtimeID)
		return p.saveRuntimeRegistry()
	})
}

func (p *Plane) Heartbeat(ctx context.Context, hb controlplane.RuntimeHeartbeat) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return p.withFileLock(ctx, func() error {
		if err := p.reloadRuntimeRegistry(); err != nil {
			return err
		}
		rec, ok := p.runtimes[hb.RuntimeID]
		if !ok {
			return planeErrorf(ErrTaskDBPlaneRuntime, "heartbeat", "heartbeat for unknown runtime %q", hb.RuntimeID)
		}
		rec.LastHeartbeat = p.now().UTC()
		applyHeartbeat(&rec.RuntimeRegistration, hb)
		p.runtimes[hb.RuntimeID] = rec
		if err := p.saveRuntimeRegistry(); err != nil {
			return err
		}
		leases, err := loadRuntimeLeaseRegistryOrEmpty(p.leasePath)
		if err != nil {
			return err
		}
		leases, changed := refreshRuntimeLeases(leases, rec, hb.RunningTaskIDs, rec.LastHeartbeat, p.leaseTTL)
		if !changed {
			return nil
		}
		return saveRuntimeLeaseRegistry(p.leasePath, p.path, leases, rec.LastHeartbeat)
	})
}

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
		if record.State == task.StateQueued {
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

func taskRequestFromRecord(path string, record taskdb.TaskRecord, provider string, prompt string, lease RuntimeLeaseRecord) bridge.TaskRequest {
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
	switch record.State {
	case task.StatePreparing, task.StateRunning, task.StateNeedsInput, task.StateBlocked, task.StateValidating, task.StatePatchReady, task.StateHumanReview:
		return db, nil
	case task.StateClaimed:
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
	switch record.State {
	case task.StateRunning, task.StateNeedsInput, task.StateBlocked, task.StateValidating, task.StatePatchReady, task.StateHumanReview:
		return db, nil
	case task.StateClaimed:
		var err error
		db, err = applyTransition(db, record, task.StatePreparing, ir.EventWorkdirPreparing, "workspace preparation started", "preparing", now)
		if err != nil {
			return taskdb.TaskDB{}, err
		}
		record, _ = findTask(db, taskID)
		fallthrough
	case task.StatePreparing:
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
	status := res.Status
	if status == "" {
		status = agentbridge.ResultCompleted
	}
	switch status {
	case agentbridge.ResultCompleted:
		if record.State == task.StateValidating || record.State == task.StatePatchReady || record.State == task.StateHumanReview {
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
		if record.State == task.StateBlocked {
			return db, nil
		}
		if record.State == task.StateClaimed {
			var err error
			db, err = ensurePreparing(db, taskID, now)
			if err != nil {
				return taskdb.TaskDB{}, err
			}
			record, _ = findTask(db, taskID)
		}
		if record.State == task.StatePreparing || record.State == task.StateRunning {
			return applyTransition(db, record, task.StateBlocked, ir.EventBlockerRaised, resultReason(res, "runtime eligibility blocked task"), "blocked", now)
		}
		return applyTransition(db, record, task.StateFailed, ir.EventTaskFailed, resultReason(res, "runtime eligibility blocked task from invalid state"), "failed:blocked-invalid-state", now)
	default:
		return applyTransition(db, record, task.StateFailed, ir.EventTaskFailed, resultReason(res, string(status)), "failed:"+string(status), now)
	}
}

func applyTransition(db taskdb.TaskDB, record taskdb.TaskRecord, to task.TaskState, event ir.EventType, reason string, commandSuffix string, now time.Time) (taskdb.TaskDB, error) {
	approvalID := approvalIDForTask(db, record.ID)
	if requiresApproval(db, record) && approvalID == "" {
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "apply-transition", "task %s requires approval_id before %s", record.ID, event)
	}
	updated, _, _, err := taskdb.ApplyGuardedTaskTransition(db, taskdb.TaskTransitionInput{
		TaskID:  record.ID,
		ToState: to,
		Event:   event,
		Actor:   defaultActor,
		Source:  sourceName,
		Reason:  reason,
		Guard:   guardFor(db, record, commandSuffix, approvalID),
	}, now)
	if err != nil {
		return taskdb.TaskDB{}, err
	}
	return updated, nil
}

func guardFor(db taskdb.TaskDB, record taskdb.TaskRecord, suffix string, approvalID string) taskdb.TaskMutationGuardInput {
	return taskdb.TaskMutationGuardInput{
		CommandID:   commandIDPrefix + record.ID + ":" + suffix,
		Provider:    providerFor(db, record),
		DecisionLLM: decisionLLMFor(db, record),
		ApprovalID:  approvalID,
	}
}

func findTask(db taskdb.TaskDB, taskID string) (taskdb.TaskRecord, bool) {
	for _, record := range db.Tasks {
		if record.ID == taskID {
			return record, true
		}
	}
	return taskdb.TaskRecord{}, false
}

func providerFor(db taskdb.TaskDB, record taskdb.TaskRecord) string {
	return textutil.FirstNonEmptyTrimmed(record.RecommendedProvider, db.RecommendedProvider)
}

func decisionLLMFor(db taskdb.TaskDB, record taskdb.TaskRecord) string {
	return textutil.FirstNonEmptyTrimmed(record.RecommendedDecisionLLM, db.RecommendedDecisionLLM)
}

func promptFor(record taskdb.TaskRecord) string {
	return textutil.FirstNonEmptyTrimmed(record.HarnessNextDirection, record.Title)
}

func requiresApproval(db taskdb.TaskDB, record taskdb.TaskRecord) bool {
	return record.RequiresHumanApproval || db.DecisionGate == "human-approval-required"
}

func approvalIDForTask(db taskdb.TaskDB, taskID string) string {
	for i := len(db.CommandReceipts) - 1; i >= 0; i-- {
		receipt := db.CommandReceipts[i]
		if receipt.TaskID == taskID && strings.TrimSpace(receipt.ApprovalID) != "" {
			return strings.TrimSpace(receipt.ApprovalID)
		}
	}
	return ""
}

func providerAvailable(db taskdb.TaskDB, provider string) bool {
	if len(db.ProviderCandidates) == 0 {
		return true
	}
	for _, candidate := range db.ProviderCandidates {
		if candidate.ID == provider {
			return candidate.Available
		}
	}
	return false
}

func (p *Plane) runtimeSelectedForTask(provider string, runtimeID string) bool {
	_, ok := p.runtimeSelectionForTask(provider, runtimeID)
	return ok
}

func (p *Plane) runtimeSelectionForTask(provider string, runtimeID string) (scheduling.RuntimeSelection, bool) {
	if len(p.runtimes) == 0 {
		return scheduling.RuntimeSelection{Runtime: scheduling.RuntimeCapability{
			RuntimeID: capability.RuntimeID(runtimeID),
			Provider:  capability.ProviderKind(provider),
			Available: true,
		}}, true
	}
	selection, ok := scheduling.SelectRuntime(scheduling.TaskRequirements{
		Provider: capability.ProviderKind(provider),
	}, p.runtimeCandidatesForProvider(provider))
	if !ok {
		return scheduling.RuntimeSelection{}, false
	}
	if string(selection.Runtime.RuntimeID) != runtimeID {
		return scheduling.RuntimeSelection{}, false
	}
	return selection, true
}

func (p *Plane) runtimeCandidatesForProvider(provider string) []scheduling.RuntimeCapability {
	ids := make([]string, 0, len(p.runtimes))
	for id := range p.runtimes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]scheduling.RuntimeCapability, 0, len(ids))
	for _, id := range ids {
		candidate, ok := runtimeCapabilityForProvider(p.runtimes[id], provider)
		if ok {
			out = append(out, candidate)
		}
	}
	return out
}

func runtimeCapabilityForProvider(rec controlplane.RegisteredRuntime, provider string) (scheduling.RuntimeCapability, bool) {
	provider = strings.TrimSpace(provider)
	if rec.RuntimeID == "" || provider == "" {
		return scheduling.RuntimeCapability{}, false
	}
	prefix := "provider." + provider + "."
	available, ok := rec.Capabilities[prefix+"available"]
	if !ok {
		return scheduling.RuntimeCapability{}, false
	}
	return scheduling.RuntimeCapability{
		RuntimeID:                 capability.RuntimeID(rec.RuntimeID),
		Provider:                  capability.ProviderKind(provider),
		CapabilityFingerprint:     capability.CapabilityFingerprint(rec.CapabilityAttributes[prefix+"capability_fingerprint"]),
		SlotLimit:                 rec.SlotLimit,
		SlotsInUse:                rec.SlotsInUse,
		Available:                 available,
		CompatibilityStatus:       capability.CompatibilityStatus(rec.CapabilityAttributes[prefix+"compatibility_status"]),
		RequiresExperimentalOptIn: rec.Capabilities[prefix+"requires_experimental_opt_in"],
		SupportsStreaming:         rec.Capabilities[prefix+"supports_streaming"],
		SupportsResume:            rec.Capabilities[prefix+"supports_resume"],
		SupportsSystem:            rec.Capabilities[prefix+"supports_system"],
		SupportsMaxTurns:          rec.Capabilities[prefix+"supports_max_turns"],
		SupportsMCP:               rec.Capabilities[prefix+"supports_mcp"],
		SupportsToolHooks:         rec.Capabilities[prefix+"supports_tool_hooks"],
		SupportsUsage:             rec.Capabilities[prefix+"supports_usage"],
		SupportsWorktree:          rec.Capabilities[prefix+"supports_worktree"],
	}, true
}

func timeoutCanOriginate(state task.TaskState) bool {
	switch state {
	case task.StateRunning, task.StateNeedsInput, task.StateBlocked, task.StateValidating, task.StateHumanReview:
		return true
	default:
		return false
	}
}

func resultReason(res agentbridge.Result, fallback string) string {
	return textutil.FirstNonEmptyTrimmed(res.Error, res.Output, fallback)
}
