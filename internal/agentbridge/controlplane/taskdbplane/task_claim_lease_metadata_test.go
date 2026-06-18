package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestClaimTaskPersistsRuntimeLeaseMetadata(t *testing.T) {
	path := writeTaskDB(t, singleQueuedCodexTaskDB())
	plane := newTestPlane(t, path)
	registerRuntimeForProvider(t, plane, "runtime-1", "codex", 2, 0)

	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req == nil {
		t.Fatal("ClaimTask returned nil request")
		return
	}
	assertClaimLeaseMetadata(t, req.Metadata)
	assertClaimLeaseRegistry(t, readRuntimeLeaseRegistry(t, plane.leasePath), path)
}

func assertClaimLeaseMetadata(t *testing.T, metadata map[string]string) {
	t.Helper()
	if metadata[controlplane.MetadataRuntimeLeaseID] != "runtime-lease:codex-task:1" {
		t.Fatalf("lease metadata missing: %+v", metadata)
	}
	if metadata[controlplane.MetadataRuntimeFencingToken] != "1" {
		t.Fatalf("fencing token metadata missing: %+v", metadata)
	}
	if metadata[controlplane.MetadataRuntimeCapabilityFingerprint] != "runtime-1-fp" {
		t.Fatalf("capability fingerprint metadata missing: %+v", metadata)
	}
}

func assertClaimLeaseRegistry(t *testing.T, registry RuntimeLeaseRegistry, path string) {
	t.Helper()
	if registry.SchemaVersion != RuntimeLeaseRegistrySchemaVersion || registry.TaskDBPath != path {
		t.Fatalf("lease registry identity mismatch: %+v", registry)
	}
	if len(registry.Leases) != 1 {
		t.Fatalf("lease count = %d, want 1: %+v", len(registry.Leases), registry.Leases)
	}
	lease := registry.Leases[0]
	if lease.TaskID != "codex-task" || lease.RuntimeID != "runtime-1" || lease.FencingToken != 1 {
		t.Fatalf("lease mismatch: %+v", lease)
	}
	if lease.CapabilityFingerprint != "runtime-1-fp" || lease.ReleasedAt != nil {
		t.Fatalf("lease fingerprint/release mismatch: %+v", lease)
	}
	if !lease.LeaseUntil.After(lease.ClaimedAt) {
		t.Fatalf("lease deadline should be after claim time: %+v", lease)
	}
}
