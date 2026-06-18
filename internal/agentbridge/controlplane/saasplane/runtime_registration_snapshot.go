package saasplane

import (
	"sort"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

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

func runtimeAvailability(rt controlplane.RuntimeRegistration, provider string) (string, string) {
	if available, ok := rt.Capabilities["provider."+provider+".available"]; ok && !available {
		return "offline", "missing"
	}
	return "online", "detected"
}
