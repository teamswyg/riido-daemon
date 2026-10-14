package taskdbplane

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func writeActiveRuntimeLease(t *testing.T, plane *Plane, taskID string) context.Context {
	t.Helper()
	now := time.Date(2026, 5, 25, 1, 0, 0, 0, time.UTC)
	return writeRuntimeLease(t, plane, RuntimeLeaseRecord{
		LeaseID:               "runtime-lease:" + taskID + ":1",
		TaskID:                taskID,
		RuntimeID:             "runtime-1",
		CapabilityFingerprint: "runtime-1-fp",
		ClaimedAt:             now,
		LeaseUntil:            now.Add(time.Hour),
		FencingToken:          1,
	})
}

func writeRuntimeLease(t *testing.T, plane *Plane, lease RuntimeLeaseRecord) context.Context {
	t.Helper()
	registry := RuntimeLeaseRegistry{Leases: []RuntimeLeaseRecord{lease}}
	if err := saveRuntimeLeaseRegistry(plane.leasePath, plane.path, registry, lease.ClaimedAt); err != nil {
		t.Fatalf("saveRuntimeLeaseRegistry: %v", err)
	}
	return contextWithRuntimeLease(lease)
}

func contextWithTaskRequest(t *testing.T, req *bridge.TaskRequest) context.Context {
	t.Helper()
	report, ok := controlplane.TaskReportContextFromMetadata(req.Metadata)
	if !ok {
		t.Fatalf("request missing task report context metadata: %+v", req.Metadata)
	}
	return controlplane.ContextWithTaskReport(context.Background(), report)
}

func contextWithRuntimeLease(lease RuntimeLeaseRecord) context.Context {
	return controlplane.ContextWithTaskReport(context.Background(), controlplane.TaskReportContext{
		RuntimeLeaseID:               lease.LeaseID,
		RuntimeFencingToken:          lease.FencingToken,
		RuntimeFencingTokenSet:       true,
		RuntimeCapabilityFingerprint: lease.CapabilityFingerprint,
	})
}
