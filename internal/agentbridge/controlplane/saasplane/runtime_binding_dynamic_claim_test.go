package saasplane

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneRenewActiveLeaseWhenRuntimeSnapshotFails(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("jykim1")}
	fake.enqueue(queuedHTTPAssignment("asn-1", "dynamic binding task"))
	ctx := t.Context()
	plane := newRuntimeBindingPlane(t, fake, nil)
	const runtimeID = "daemon-1:codex"
	req, err := plane.ClaimTask(ctx, runtimeID)
	if err != nil || req == nil {
		t.Fatalf("ClaimTask = %+v, %v", req, err)
	}
	if err := plane.StartTask(ctx, req.ID); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.RegisterRuntime(ctx, controlplane.RuntimeRegistration{RuntimeID: runtimeID, Provider: "codex"}); err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}
	if err := plane.withState(ctx, func(s *planeState) { s.lastRuntimeSnapshotSync = time.Time{} }); err != nil {
		t.Fatalf("reset runtime snapshot interval: %v", err)
	}
	fake.failNext("/v1/daemon/runtime-snapshot", 1, http.StatusBadRequest)
	err = plane.Heartbeat(ctx, controlplane.RuntimeHeartbeat{RuntimeID: runtimeID, RunningTaskIDs: []string{req.ID}})
	var status httpStatusError
	if !errors.As(err, &status) || status.StatusCode != http.StatusBadRequest {
		t.Fatalf("Heartbeat error = %v, want the snapshot rejection", err)
	}
	if len(fake.heartbeats) != 1 || len(fake.heartbeats[0].ActiveAssignmentIDs) != 1 || fake.heartbeats[0].ActiveAssignmentIDs[0] != req.ID {
		t.Fatalf("snapshot rejection skipped the active lease heartbeat: %+v", fake.heartbeats)
	}
}

func TestPlaneClaimsDynamicAgentBinding(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("jykim1")}
	fake.enqueue(queuedHTTPAssignment("asn-1", "dynamic binding task"))
	plane := newRuntimeBindingPlane(t, fake, nil)

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req == nil || req.ID != "asn-1" || req.Provider != "codex" || req.Metadata[MetadataAgentID] != "jykim1" {
		t.Fatalf("dynamic claim = %+v", req)
	}
	if err := plane.StartTask(context.Background(), req.ID); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	err = plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      "daemon-1:codex",
		RunningTaskIDs: []string{req.ID},
	})
	if err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	assertDynamicHeartbeat(t, fake, req.ID)
	err = plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{
		Kind: agentbridge.EventProgress,
		Text: "dynamic progress",
	})
	if err != nil {
		t.Fatalf("ReportEvent: %v", err)
	}
	if len(fake.events) < 2 || fake.events[len(fake.events)-1].RuntimeID != "daemon-1:codex" {
		t.Fatalf("dynamic events = %+v", fake.events)
	}
}
