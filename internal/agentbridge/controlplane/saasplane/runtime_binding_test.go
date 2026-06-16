package saasplane

import (
	"context"
	"strings"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneHeartbeatRefreshesAggregatedRuntimeSnapshot(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	startedAt := time.Now().Add(-5 * time.Minute).UTC()
	plane, err := New(Config{
		BaseURL:      fake.URL(),
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
		Profile:      "development",
		AppVersion:   "v0.0.13",
		PID:          8765,
		StartedAt:    startedAt,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	if err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		DaemonID:   "daemon-1",
		RuntimeID:  "daemon-1:codex",
		Provider:   "codex",
		DeviceName: "주윤의 MacBook",
		Models: []controlplane.RuntimeModel{
			{ModelID: "gpt-5.5", Label: "gpt-5.5", IsDefault: true},
		},
		Capabilities: map[string]bool{
			"provider.codex.requires_experimental_opt_in": true,
		},
		CapabilityAttributes: map[string]string{
			"provider.codex.provider_version": "codex-cli 0.133.0",
		},
	}); err != nil {
		t.Fatalf("RegisterRuntime codex: %v", err)
	}
	if err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		DaemonID:   "daemon-1",
		RuntimeID:  "daemon-1:cursor",
		Provider:   "cursor",
		DeviceName: "주윤의 MacBook",
	}); err != nil {
		t.Fatalf("RegisterRuntime cursor: %v", err)
	}
	if len(fake.runtimeSnapshots) != 2 {
		t.Fatalf("registration snapshots = %+v", fake.runtimeSnapshots)
	}

	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID: "daemon-1:codex",
	}); err != nil {
		t.Fatalf("Heartbeat codex: %v", err)
	}
	if len(fake.runtimeSnapshots) != 3 {
		t.Fatalf("heartbeat should append one aggregated snapshot, got %+v", fake.runtimeSnapshots)
	}
	snapshot := fake.runtimeSnapshots[2]
	if snapshot.DaemonID != "daemon-1" || snapshot.DeviceID != "device-1" || snapshot.DeviceDisplayName != "주윤의 MacBook" {
		t.Fatalf("heartbeat snapshot identity = %+v", snapshot)
	}
	if snapshot.Profile != "development" || snapshot.AppVersion != "v0.0.13" || snapshot.PID != 8765 || !snapshot.StartedAt.Equal(startedAt) || snapshot.UptimeSeconds <= 0 {
		t.Fatalf("heartbeat daemon facts = %+v", snapshot)
	}
	if len(snapshot.Runtimes) != 2 ||
		snapshot.Runtimes[0].RuntimeID != "daemon-1:codex" ||
		snapshot.Runtimes[1].RuntimeID != "daemon-1:cursor" {
		t.Fatalf("heartbeat snapshot must aggregate sorted runtimes: %+v", snapshot.Runtimes)
	}
	if len(snapshot.Runtimes[0].Models) != 1 || snapshot.Runtimes[0].Models[0].ModelID != "gpt-5.5" || snapshot.Runtimes[0].ProviderVersion != "codex-cli 0.133.0" || !snapshot.Runtimes[0].RequiresExperimentalOptIn {
		t.Fatalf("codex runtime facts lost in heartbeat snapshot: %+v", snapshot.Runtimes[0])
	}

	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID: "daemon-1:cursor",
	}); err != nil {
		t.Fatalf("Heartbeat cursor: %v", err)
	}
	if len(fake.runtimeSnapshots) != 3 {
		t.Fatalf("same heartbeat window should not create per-runtime snapshot fanout: %+v", fake.runtimeSnapshots)
	}
}

func TestPlaneClaimsDynamicAgentBinding(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{{
		AgentID:         "jykim1",
		DaemonID:        "daemon-1",
		DeviceID:        "device-1",
		RuntimeID:       "daemon-1:codex",
		RuntimeProvider: "codex",
	}}
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "dynamic binding task",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	plane, err := New(Config{
		BaseURL:      fake.URL(),
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

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
	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      "daemon-1:codex",
		RunningTaskIDs: []string{req.ID},
	}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	if len(fake.heartbeats) != 1 || fake.heartbeats[0].RuntimeID != "daemon-1:codex" || len(fake.heartbeats[0].ActiveAssignmentIDs) != 1 || fake.heartbeats[0].ActiveAssignmentIDs[0] != "asn-1" {
		t.Fatalf("dynamic heartbeats = %+v", fake.heartbeats)
	}
	if err := plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{Kind: agentbridge.EventProgress, Text: "dynamic progress"}); err != nil {
		t.Fatalf("ReportEvent: %v", err)
	}
	if len(fake.events) < 2 || fake.events[len(fake.events)-1].RuntimeID != "daemon-1:codex" {
		t.Fatalf("dynamic events = %+v", fake.events)
	}
}

func TestPlaneCachesDynamicAgentBindingsAcrossClaimWave(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{
		{
			AgentID:         "agent-codex",
			DaemonID:        "daemon-1",
			DeviceID:        "device-1",
			RuntimeID:       "daemon-1:codex",
			RuntimeProvider: "codex",
		},
		{
			AgentID:         "agent-cursor",
			DaemonID:        "daemon-1",
			DeviceID:        "device-1",
			RuntimeID:       "daemon-1:cursor",
			RuntimeProvider: "cursor",
		},
	}
	plane, err := New(Config{
		BaseURL:      fake.URL(),
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	for _, runtimeID := range []string{"daemon-1:codex", "daemon-1:cursor"} {
		req, err := plane.ClaimTask(context.Background(), runtimeID)
		if err != nil {
			t.Fatalf("ClaimTask %s: %v", runtimeID, err)
		}
		if req != nil {
			t.Fatalf("empty queue should not claim task: %+v", req)
		}
	}
	if got := fake.requestCount("/v1/daemon/agent-bindings"); got != 1 {
		t.Fatalf("agent-bindings request count = %d, want 1", got)
	}
	if got := len(fake.pollRequestsFor("agent-codex")); got != 1 {
		t.Fatalf("agent-codex poll count = %d, want 1", got)
	}
	if got := len(fake.pollRequestsFor("agent-cursor")); got != 1 {
		t.Fatalf("agent-cursor poll count = %d, want 1", got)
	}
	if err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "daemon-1:codex",
		Provider:  "codex",
	}); err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}
	if _, err := plane.ClaimTask(context.Background(), "daemon-1:codex"); err != nil {
		t.Fatalf("ClaimTask after runtime snapshot: %v", err)
	}
	if got := fake.requestCount("/v1/daemon/agent-bindings"); got != 2 {
		t.Fatalf("agent-bindings request count after runtime snapshot = %d, want 2", got)
	}
}

func TestPlaneCachesEmptyDynamicAgentBindingsAcrossClaimWave(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	plane, err := New(Config{
		BaseURL:      fake.URL(),
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	for _, runtimeID := range []string{"daemon-1:codex", "daemon-1:cursor"} {
		req, err := plane.ClaimTask(context.Background(), runtimeID)
		if err != nil {
			t.Fatalf("ClaimTask %s: %v", runtimeID, err)
		}
		if req != nil {
			t.Fatalf("empty binding list should not claim task: %+v", req)
		}
	}
	if got := fake.requestCount("/v1/daemon/agent-bindings"); got != 1 {
		t.Fatalf("empty agent-bindings request count = %d, want 1", got)
	}
}

func TestPlaneRejectsMissingBearerToken(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bearerToken = "secret"
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	_, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected 401 without token, got %v", err)
	}
}

func TestPlaneUsesSharedAssignmentContractSurface(t *testing.T) {
	if assignmentcontract.SchemaVersion != "riido-ai-server.v1" {
		t.Fatalf("schema version = %q", assignmentcontract.SchemaVersion)
	}
	if !assignmentcontract.PollStart.Valid() || !assignmentcontract.AssignmentReady.Valid() {
		t.Fatal("shared assignment contract validation is not wired")
	}
}

func newTestPlane(t *testing.T, baseURL string, agents []AgentBinding) *Plane {
	t.Helper()
	return newTestPlaneWithToken(t, baseURL, agents, "")
}
