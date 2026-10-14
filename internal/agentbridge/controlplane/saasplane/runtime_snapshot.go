package saasplane

import (
	"context"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func runtimeSnapshotFromHeartbeat(hb controlplane.RuntimeHeartbeat) (RuntimeSnapshotRecord, bool) {
	runtimeID := strings.TrimSpace(hb.RuntimeID)
	provider := providerFromRuntimeID(runtimeID)
	if runtimeID == "" || provider == "" {
		return RuntimeSnapshotRecord{}, false
	}
	return RuntimeSnapshotRecord{
		RuntimeID:      runtimeID,
		Kind:           runtimeKindForProvider(provider),
		Availability:   "online",
		DetectionState: "detected",
	}, true
}

func (p *Plane) postRuntimeSnapshot(ctx context.Context, runtimes []RuntimeSnapshotRecord, deviceName string) error {
	var out struct {
		SchemaVersion string `json:"schema_version"`
	}
	err := p.postJSON(ctx, "/v1/daemon/runtime-snapshot", DeviceRuntimeSnapshotSyncRequest{
		DaemonID:          p.cfg.DaemonID,
		DeviceID:          p.cfg.DeviceID,
		DeviceDisplayName: textutil.FirstNonEmptyTrimmed(deviceName, p.cfg.DeviceID),
		Profile:           p.cfg.Profile,
		AppVersion:        p.cfg.AppVersion,
		PID:               p.cfg.PID,
		UptimeSeconds:     p.daemonUptimeSeconds(),
		StartedAt:         p.cfg.StartedAt,
		Runtimes:          runtimes,
	}, &out)
	if err != nil {
		return err
	}
	p.invalidateAgentBindingsCache(ctx)
	return nil
}

func (p *Plane) daemonUptimeSeconds() int64 {
	if p.cfg.StartedAt.IsZero() {
		return 0
	}
	seconds := int64(time.Since(p.cfg.StartedAt).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}
