package taskdbplane

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
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
