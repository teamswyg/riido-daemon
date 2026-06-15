package saasplane

import (
	"context"
	"net/url"
	"sort"
	"strings"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func (p *Plane) RegisterRuntime(ctx context.Context, rt controlplane.RuntimeRegistration) error {
	if !p.dynamicBindingsEnabled() {
		return nil
	}
	snapshot, deviceName, ok := runtimeSnapshotFromRegistration(rt)
	if !ok {
		return nil
	}
	var runtimes []RuntimeSnapshotRecord
	var postDeviceName string
	if err := p.withState(ctx, func(s *planeState) {
		s.registeredRuntimes[snapshot.RuntimeID] = snapshot
		if deviceName != "" {
			s.registeredDeviceName = deviceName
		}
		postDeviceName = s.registeredDeviceName
		runtimes = sortedRuntimeSnapshots(s.registeredRuntimes)
	}); err != nil {
		return err
	}
	// Post the full accumulated provider set, not just this one runtime, so the
	// control-plane device projection always reflects every known runtime —
	// undetected providers stay present as detection_state=missing instead of
	// being clobbered to an empty list under snapshot replace semantics.
	return p.postRuntimeSnapshot(ctx, runtimes, postDeviceName)
}

func sortedRuntimeSnapshots(in map[string]RuntimeSnapshotRecord) []RuntimeSnapshotRecord {
	out := make([]RuntimeSnapshotRecord, 0, len(in))
	for _, runtime := range in {
		out = append(out, runtime)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].RuntimeID < out[j].RuntimeID
	})
	return out
}

func runtimeSnapshotFromRegistration(rt controlplane.RuntimeRegistration) (RuntimeSnapshotRecord, string, bool) {
	runtimeID := strings.TrimSpace(rt.RuntimeID)
	provider := providerFromRuntimeID(textutil.FirstNonEmptyTrimmed(rt.Provider, runtimeID))
	if runtimeID == "" || provider == "" {
		return RuntimeSnapshotRecord{}, "", false
	}
	availability, detectionState := runtimeAvailability(rt, provider)
	return RuntimeSnapshotRecord{
		RuntimeID:                 runtimeID,
		Kind:                      runtimeKindForProvider(provider),
		Availability:              availability,
		DetectionState:            detectionState,
		ProviderVersion:           runtimeProviderVersion(rt, provider),
		RequiresExperimentalOptIn: runtimeRequiresExperimentalOptIn(rt, provider),
		Models:                    runtimeModels(rt.Models),
	}, strings.TrimSpace(rt.DeviceName), true
}

func (p *Plane) DeregisterRuntime(context.Context, string) error {
	return nil
}

func runtimeAvailability(rt controlplane.RuntimeRegistration, provider string) (string, string) {
	if available, ok := rt.Capabilities["provider."+provider+".available"]; ok && !available {
		return "offline", "missing"
	}
	return "online", "detected"
}

func runtimeModels(in []controlplane.RuntimeModel) []RuntimeModelRecord {
	out := make([]RuntimeModelRecord, 0, len(in))
	for _, model := range in {
		out = append(out, RuntimeModelRecord{
			ModelID:   model.ModelID,
			Label:     model.Label,
			IsDefault: model.IsDefault,
		})
	}
	return out
}

func runtimeRequiresExperimentalOptIn(rt controlplane.RuntimeRegistration, provider string) bool {
	if len(rt.Capabilities) == 0 {
		return false
	}
	key := "provider." + provider + ".requires_experimental_opt_in"
	return rt.Capabilities[key]
}

func runtimeProviderVersion(rt controlplane.RuntimeRegistration, provider string) string {
	return strings.TrimSpace(rt.CapabilityAttributes["provider."+provider+".provider_version"])
}

func (p *Plane) Heartbeat(ctx context.Context, hb controlplane.RuntimeHeartbeat) error {
	if p.dynamicBindingsEnabled() {
		if err := p.refreshRegisteredRuntimeSnapshot(ctx, hb); err != nil {
			return err
		}
		assignmentsByAgent, err := p.activeAssignmentsByAgentForHeartbeat(ctx, hb.RunningTaskIDs)
		if err != nil {
			return err
		}
		for agentID, assignmentIDs := range assignmentsByAgent {
			if len(assignmentIDs) == 0 {
				continue
			}
			var out assignmentcontract.AgentHeartbeatResponse
			if err := p.postJSON(ctx, "/v1/agents/"+url.PathEscape(agentID)+"/heartbeat", assignmentcontract.AgentHeartbeatRequest{
				DaemonID:            p.cfg.DaemonID,
				DeviceID:            p.cfg.DeviceID,
				RuntimeID:           hb.RuntimeID,
				RunningTaskIDs:      append([]string(nil), hb.RunningTaskIDs...),
				ActiveAssignmentIDs: assignmentIDs,
			}, &out); err != nil {
				return err
			}
			if err := p.deliverUnrefreshedHeartbeatCancels(ctx, assignmentIDs, out); err != nil {
				return err
			}
		}
		return nil
	}
	agentID, ok := agentFromRuntimeID(hb.RuntimeID)
	if !ok {
		return nil
	}
	assignmentIDs, err := p.activeAssignmentIDsForHeartbeat(ctx, agentID, hb.RunningTaskIDs)
	if err != nil {
		return err
	}
	if len(assignmentIDs) == 0 {
		return nil
	}
	var out assignmentcontract.AgentHeartbeatResponse
	if err := p.postJSON(ctx, "/v1/agents/"+url.PathEscape(agentID)+"/heartbeat", assignmentcontract.AgentHeartbeatRequest{
		DaemonID:            p.cfg.DaemonID,
		DeviceID:            p.cfg.DeviceID,
		RuntimeID:           hb.RuntimeID,
		RunningTaskIDs:      append([]string(nil), hb.RunningTaskIDs...),
		ActiveAssignmentIDs: assignmentIDs,
	}, &out); err != nil {
		return err
	}
	return p.deliverUnrefreshedHeartbeatCancels(ctx, assignmentIDs, out)
}

func (p *Plane) refreshRegisteredRuntimeSnapshot(ctx context.Context, hb controlplane.RuntimeHeartbeat) error {
	now := time.Now()
	var runtimes []RuntimeSnapshotRecord
	var deviceName string
	err := p.withState(ctx, func(s *planeState) {
		if len(s.registeredRuntimes) == 0 {
			fallback, ok := runtimeSnapshotFromHeartbeat(hb)
			if !ok {
				return
			}
			s.registeredRuntimes[fallback.RuntimeID] = fallback
		}
		if !s.lastRuntimeSnapshotSync.IsZero() && now.Sub(s.lastRuntimeSnapshotSync) < runtimeSnapshotHeartbeatMinInterval {
			return
		}
		s.lastRuntimeSnapshotSync = now
		deviceName = s.registeredDeviceName
		runtimes = sortedRuntimeSnapshots(s.registeredRuntimes)
	})
	if err != nil || len(runtimes) == 0 {
		return err
	}
	return p.postRuntimeSnapshot(ctx, runtimes, deviceName)
}
