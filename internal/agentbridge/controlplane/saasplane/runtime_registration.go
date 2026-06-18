package saasplane

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (p *Plane) RegisterRuntime(ctx context.Context, rt controlplane.RuntimeRegistration) error {
	if !p.dynamicBindingsEnabled() {
		return nil
	}
	snapshot, deviceName, ok := runtimeSnapshotFromRegistration(rt)
	if !ok {
		return nil
	}
	runtimes, postDeviceName, err := p.recordRuntimeSnapshot(ctx, snapshot, deviceName)
	if err != nil {
		return err
	}
	return p.postRuntimeSnapshot(ctx, runtimes, postDeviceName)
}

func (p *Plane) recordRuntimeSnapshot(ctx context.Context, snapshot RuntimeSnapshotRecord, deviceName string) ([]RuntimeSnapshotRecord, string, error) {
	var runtimes []RuntimeSnapshotRecord
	var postDeviceName string
	err := p.withState(ctx, func(s *planeState) {
		s.registeredRuntimes[snapshot.RuntimeID] = snapshot
		if deviceName != "" {
			s.registeredDeviceName = deviceName
		}
		postDeviceName = s.registeredDeviceName
		runtimes = sortedRuntimeSnapshots(s.registeredRuntimes)
	})
	return runtimes, postDeviceName, err
}

func (p *Plane) DeregisterRuntime(context.Context, string) error {
	return nil
}
