package taskdbplane

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	c9lock "github.com/teamswyg/riido-daemon/internal/lock"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
)

func (p *Plane) saveRuntimeRegistry() error {
	registry := RuntimeRegistry{
		SchemaVersion: RuntimeRegistrySchemaVersion,
		TaskDBPath:    p.path,
		UpdatedAt:     p.now().UTC(),
		Runtimes:      sortedRuntimeRegistry(p.runtimes),
	}
	return writeJSONAtomic(p.registryPath, registry)
}

func (p *Plane) reloadRuntimeRegistry() error {
	runtimes, err := loadRuntimeRegistryOrEmpty(p.registryPath)
	if err != nil {
		return err
	}
	p.runtimes = runtimes
	return nil
}

func sortedRuntimeRegistry(runtimes map[string]controlplane.RegisteredRuntime) []controlplane.RegisteredRuntime {
	ids := make([]string, 0, len(runtimes))
	for id := range runtimes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]controlplane.RegisteredRuntime, 0, len(ids))
	for _, id := range ids {
		out = append(out, runtimes[id])
	}
	return out
}

func loadRuntimeRegistryOrEmpty(path string) (map[string]controlplane.RegisteredRuntime, error) {
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]controlplane.RegisteredRuntime{}, nil
	}
	if err != nil {
		return nil, planeWrapf(ErrTaskDBPlaneRegistry, "registry.load", err, "read runtime registry")
	}
	var registry RuntimeRegistry
	if err := json.Unmarshal(body, &registry); err != nil {
		return nil, planeWrapf(ErrTaskDBPlaneRegistry, "registry.load", err, "decode runtime registry")
	}
	if registry.SchemaVersion != RuntimeRegistrySchemaVersion {
		return nil, planeErrorf(ErrTaskDBPlaneRegistry, "registry.load", "runtime registry schema mismatch: got %q want %q", registry.SchemaVersion, RuntimeRegistrySchemaVersion)
	}
	out := make(map[string]controlplane.RegisteredRuntime, len(registry.Runtimes))
	for _, runtime := range registry.Runtimes {
		if runtime.RuntimeID != "" {
			out[runtime.RuntimeID] = runtime
		}
	}
	return out, nil
}

func applyHeartbeat(reg *controlplane.RuntimeRegistration, hb controlplane.RuntimeHeartbeat) {
	if hb.RuntimeID != "" {
		reg.RuntimeID = hb.RuntimeID
	}
	if hb.DeviceName != "" {
		reg.DeviceName = hb.DeviceName
	}
	reg.UptimeSeconds = hb.UptimeSeconds
	reg.SlotLimit = hb.SlotLimit
	reg.SlotsInUse = hb.SlotsInUse
	reg.RunningTaskIDs = append([]string(nil), hb.RunningTaskIDs...)
	sort.Strings(reg.RunningTaskIDs)
}

func loadRuntimeLeaseRegistryOrEmpty(path string) (RuntimeLeaseRegistry, error) {
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return RuntimeLeaseRegistry{
			SchemaVersion: RuntimeLeaseRegistrySchemaVersion,
			Leases:        []RuntimeLeaseRecord{},
		}, nil
	}
	if err != nil {
		return RuntimeLeaseRegistry{}, planeWrapf(ErrTaskDBPlaneRegistry, "lease-registry.load", err, "read runtime lease registry")
	}
	var registry RuntimeLeaseRegistry
	if err := json.Unmarshal(body, &registry); err != nil {
		return RuntimeLeaseRegistry{}, planeWrapf(ErrTaskDBPlaneRegistry, "lease-registry.load", err, "decode runtime lease registry")
	}
	if registry.SchemaVersion != RuntimeLeaseRegistrySchemaVersion {
		return RuntimeLeaseRegistry{}, planeErrorf(ErrTaskDBPlaneRegistry, "lease-registry.load", "runtime lease registry schema mismatch: got %q want %q", registry.SchemaVersion, RuntimeLeaseRegistrySchemaVersion)
	}
	if registry.Leases == nil {
		registry.Leases = []RuntimeLeaseRecord{}
	}
	return registry, nil
}

func saveRuntimeLeaseRegistry(path, taskDBPath string, registry RuntimeLeaseRegistry, now time.Time) error {
	registry.SchemaVersion = RuntimeLeaseRegistrySchemaVersion
	registry.TaskDBPath = taskDBPath
	registry.UpdatedAt = now.UTC()
	sort.Slice(registry.Leases, func(i, j int) bool {
		if registry.Leases[i].TaskID != registry.Leases[j].TaskID {
			return registry.Leases[i].TaskID < registry.Leases[j].TaskID
		}
		return registry.Leases[i].FencingToken < registry.Leases[j].FencingToken
	})
	return writeJSONAtomic(path, registry)
}

func acquireRuntimeLease(registry RuntimeLeaseRegistry, taskID, runtimeID, capabilityFingerprint string, now time.Time, ttl time.Duration) (RuntimeLeaseRegistry, RuntimeLeaseRecord, bool) {
	if ttl <= 0 {
		ttl = defaultRuntimeLeaseTTL
	}
	idx := runtimeLeaseIndex(registry.Leases, taskID)
	if idx >= 0 {
		existing := registry.Leases[idx]
		if existing.isActive(now) {
			if existing.RuntimeID != runtimeID || existing.CapabilityFingerprint != capabilityFingerprint {
				return registry, RuntimeLeaseRecord{}, false
			}
			existing.LeaseUntil = now.Add(ttl)
			registry.Leases[idx] = existing
			return registry, existing, true
		}
		lease := RuntimeLeaseRecord{
			LeaseID:               runtimeLeaseID(taskID, existing.FencingToken+1),
			TaskID:                taskID,
			RuntimeID:             runtimeID,
			CapabilityFingerprint: capabilityFingerprint,
			ClaimedAt:             now,
			LeaseUntil:            now.Add(ttl),
			FencingToken:          existing.FencingToken + 1,
		}
		registry.Leases[idx] = lease
		return registry, lease, true
	}
	lease := RuntimeLeaseRecord{
		LeaseID:               runtimeLeaseID(taskID, 1),
		TaskID:                taskID,
		RuntimeID:             runtimeID,
		CapabilityFingerprint: capabilityFingerprint,
		ClaimedAt:             now,
		LeaseUntil:            now.Add(ttl),
		FencingToken:          1,
	}
	registry.Leases = append(registry.Leases, lease)
	return registry, lease, true
}

func releaseRuntimeLease(registry RuntimeLeaseRegistry, taskID string, now time.Time) (RuntimeLeaseRegistry, bool) {
	idx := runtimeLeaseIndex(registry.Leases, taskID)
	if idx < 0 {
		return registry, false
	}
	if registry.Leases[idx].ReleasedAt != nil {
		return registry, false
	}
	releasedAt := now.UTC()
	registry.Leases[idx].ReleasedAt = &releasedAt
	return registry, true
}

func refreshRuntimeLeases(registry RuntimeLeaseRegistry, rec controlplane.RegisteredRuntime, taskIDs []string, now time.Time, ttl time.Duration) (RuntimeLeaseRegistry, bool) {
	if ttl <= 0 {
		ttl = defaultRuntimeLeaseTTL
	}
	changed := false
	for _, taskID := range normalizedTaskIDs(taskIDs) {
		idx := runtimeLeaseIndex(registry.Leases, taskID)
		if idx < 0 {
			continue
		}
		lease := registry.Leases[idx]
		if !lease.isActive(now) || lease.RuntimeID != rec.RuntimeID {
			continue
		}
		if !runtimeHasCapabilityFingerprint(rec, lease.CapabilityFingerprint) {
			continue
		}
		lease.LeaseUntil = now.Add(ttl)
		registry.Leases[idx] = lease
		changed = true
	}
	return registry, changed
}

func reconcileExpiredRuntimeLeases(db taskdb.TaskDB, registry RuntimeLeaseRegistry, now time.Time) (taskdb.TaskDB, RuntimeLeaseRegistry, bool, error) {
	changed := false
	for _, lease := range append([]RuntimeLeaseRecord(nil), registry.Leases...) {
		if lease.ReleasedAt != nil || !lease.isExpired(now) {
			continue
		}
		record, ok := findTask(db, lease.TaskID)
		if !ok {
			var released bool
			registry, released = releaseRuntimeLease(registry, lease.TaskID, now)
			changed = changed || released
			continue
		}
		switch record.State.Code() {
		case task.TaskStateCodePreparing, task.TaskStateCodeRunning:
			updated, err := applyExpiredRuntimeHandoff(db, record, lease, now)
			if err != nil {
				return taskdb.TaskDB{}, RuntimeLeaseRegistry{}, false, err
			}
			db = updated
			registry, _ = releaseRuntimeLease(registry, lease.TaskID, now)
			changed = true
		case task.TaskStateCodeClaimed:
			updated, err := applyTransition(db, record, task.StateFailed, ir.EventTaskFailed, "runtime lease expired before provider execution", "lease-expired:"+lease.LeaseID+":failed", now)
			if err != nil {
				return taskdb.TaskDB{}, RuntimeLeaseRegistry{}, false, err
			}
			db = updated
			registry, _ = releaseRuntimeLease(registry, lease.TaskID, now)
			changed = true
		case task.TaskStateCodeNeedsInput:
			updated, err := applyTransition(db, record, task.StateTimedOut, ir.EventTaskTimedOut, "runtime lease expired while waiting for input", "lease-expired:"+lease.LeaseID+":timed-out", now)
			if err != nil {
				return taskdb.TaskDB{}, RuntimeLeaseRegistry{}, false, err
			}
			db = updated
			registry, _ = releaseRuntimeLease(registry, lease.TaskID, now)
			changed = true
		case task.TaskStateCodeUnknown, task.TaskStateCodeQueued, task.TaskStateCodeCreated, task.TaskStateCodeBlocked, task.TaskStateCodeValidating, task.TaskStateCodePatchReady, task.TaskStateCodeHumanReview, task.TaskStateCodeReworkQueued, task.TaskStateCodeCompleted, task.TaskStateCodeFailed, task.TaskStateCodeCancelled, task.TaskStateCodeTimedOut:
			var released bool
			registry, released = releaseRuntimeLease(registry, lease.TaskID, now)
			changed = changed || released
		}
	}
	return db, registry, changed, nil
}

func applyExpiredRuntimeHandoff(db taskdb.TaskDB, record taskdb.TaskRecord, lease RuntimeLeaseRecord, now time.Time) (taskdb.TaskDB, error) {
	updated, err := applyTransition(db, record, task.StateBlocked, ir.EventBlockerRaised, "runtime lease expired; requeue for another runtime", "lease-expired:"+lease.LeaseID+":blocked", now)
	if err != nil {
		return taskdb.TaskDB{}, err
	}
	blocked, ok := findTask(updated, record.ID)
	if !ok {
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "lease.expire-handoff", "task %s not found after lease expiry block", record.ID)
	}
	return applyTransition(updated, blocked, task.StateQueued, ir.EventBlockerResolvedRequeue, "runtime lease expired; handoff queued", "lease-expired:"+lease.LeaseID+":requeue", now)
}

func requireActiveRuntimeLease(registry RuntimeLeaseRegistry, taskID string, now time.Time, report controlplane.TaskReportContext) (RuntimeLeaseRecord, error) {
	idx := runtimeLeaseIndex(registry.Leases, taskID)
	if idx < 0 {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s has no runtime lease", taskID)
	}
	lease := registry.Leases[idx]
	if !lease.isActive(now) {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s runtime lease is not active", taskID)
	}
	if report.RuntimeLeaseID == "" {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s report missing runtime lease id", taskID)
	}
	if lease.LeaseID != report.RuntimeLeaseID {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s runtime lease id mismatch", taskID)
	}
	if !report.RuntimeFencingTokenSet {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s report missing runtime fencing token", taskID)
	}
	if lease.FencingToken != report.RuntimeFencingToken {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s runtime fencing token mismatch", taskID)
	}
	if report.RuntimeCapabilityFingerprint != "" && lease.CapabilityFingerprint != report.RuntimeCapabilityFingerprint {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s runtime capability fingerprint mismatch", taskID)
	}
	return lease, nil
}

func (r RuntimeLeaseRecord) isActive(now time.Time) bool {
	return r.ReleasedAt == nil && !r.LeaseUntil.IsZero() && !now.After(r.LeaseUntil)
}

func (r RuntimeLeaseRecord) isExpired(now time.Time) bool {
	return r.ReleasedAt == nil && !r.LeaseUntil.IsZero() && now.After(r.LeaseUntil)
}

func runtimeLeaseIndex(leases []RuntimeLeaseRecord, taskID string) int {
	for i, lease := range leases {
		if lease.TaskID == taskID {
			return i
		}
	}
	return -1
}

func runtimeLeaseID(taskID string, fencingToken int64) string {
	return fmt.Sprintf("runtime-lease:%s:%d", taskID, fencingToken)
}

func runtimeHasCapabilityFingerprint(rec controlplane.RegisteredRuntime, fingerprint string) bool {
	if strings.TrimSpace(fingerprint) == "" {
		return true
	}
	for key, value := range rec.CapabilityAttributes {
		if strings.HasPrefix(key, "provider.") && strings.HasSuffix(key, ".capability_fingerprint") && value == fingerprint {
			return true
		}
	}
	return false
}

func normalizedTaskIDs(ids []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

func taskDBChanged(before, after taskdb.TaskDB, taskID string) bool {
	if len(before.Transitions) != len(after.Transitions) || len(before.CommandReceipts) != len(after.CommandReceipts) {
		return true
	}
	beforeRecord, beforeOK := findTask(before, taskID)
	afterRecord, afterOK := findTask(after, taskID)
	if beforeOK != afterOK {
		return true
	}
	if !beforeOK {
		return false
	}
	return beforeRecord.State != afterRecord.State ||
		beforeRecord.UpdatedAt != afterRecord.UpdatedAt ||
		beforeRecord.TransitionCount != afterRecord.TransitionCount ||
		beforeRecord.CommandReceiptCount != afterRecord.CommandReceiptCount
}

func runtimeRegistryPath(taskDBPath string) string {
	if before, ok := strings.CutSuffix(taskDBPath, ".json"); ok {
		return before + ".runtimes.json"
	}
	return taskDBPath + ".runtimes.json"
}

func runtimeLeaseRegistryPath(taskDBPath string) string {
	if before, ok := strings.CutSuffix(taskDBPath, ".json"); ok {
		return before + ".leases.json"
	}
	return taskDBPath + ".leases.json"
}

func (p *Plane) withFileLock(ctx context.Context, fn func() error) error {
	return c9lock.WithFile(ctx, p.lockPath, fn)
}

func writeJSONAtomic(path string, value any) error {
	if err := fileutil.WriteJSONAtomic(path, value); err != nil {
		return planeWrapf(ErrTaskDBPlanePersistence, "json.write", err, "write JSON %s", path)
	}
	return nil
}
