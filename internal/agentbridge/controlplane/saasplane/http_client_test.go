package saasplane

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneSendsDeviceCredentialHeaders(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "hello",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	plane, err := New(Config{
		BaseURL:      fake.URL(),
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
		Agents:       []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask with device credential: %v", err)
	}
	if req == nil || req.ID != "asn-1" {
		t.Fatalf("request = %+v", req)
	}
}

func TestPlaneRetriesTransientPoll(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.failNext("/v1/agents/jykim1/poll", 1, http.StatusServiceUnavailable)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "hello",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask should retry transient poll: %v", err)
	}
	if req == nil || req.ID != "asn-1" {
		t.Fatalf("request = %+v", req)
	}
	if got := fake.requestCount("/v1/agents/jykim1/poll"); got != 2 {
		t.Fatalf("poll request count = %d, want 2", got)
	}
}

func TestPlaneRetriesTransientPollTransportError(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-transport",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "hello",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	transport := &transientTransport{
		failures: 1,
		next:     fake.server.Client().Transport,
	}
	plane, err := New(Config{
		BaseURL:    fake.URL(),
		DaemonID:   "daemon-1",
		DeviceID:   "device-1",
		Agents:     []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}},
		HTTPClient: &http.Client{Transport: transport},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask should retry transient transport error: %v", err)
	}
	if req == nil || req.ID != "asn-transport" {
		t.Fatalf("request = %+v", req)
	}
	if got := fake.requestCount("/v1/agents/jykim1/poll"); got != 1 {
		t.Fatalf("server poll request count = %d, want 1 after one client-side transport failure", got)
	}
	if transport.failures != 0 {
		t.Fatalf("transport failures remaining = %d, want 0", transport.failures)
	}
}

func TestPlaneRetriesTransientAgentBindings(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	fake.failNext("/v1/daemon/agent-bindings", 1, http.StatusBadGateway)
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
		t.Fatalf("ClaimTask should retry transient agent-bindings: %v", err)
	}
	if req == nil || req.ID != "asn-1" {
		t.Fatalf("request = %+v", req)
	}
	if got := fake.requestCount("/v1/daemon/agent-bindings"); got != 2 {
		t.Fatalf("agent-bindings request count = %d, want 2", got)
	}
}

type transientTransport struct {
	failures int
	next     http.RoundTripper
}

func (t *transientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.failures > 0 {
		t.failures--
		return nil, errors.New("temporary transport failure")
	}
	return t.next.RoundTrip(req)
}

func TestPlaneRetriesTransientHeartbeat(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.failNext("/v1/agents/jykim1/heartbeat", 1, http.StatusGatewayTimeout)
	agent := AgentBinding{AgentID: "jykim1", RuntimeProvider: "codex"}
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "hello",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	plane := newTestPlane(t, fake.URL(), []AgentBinding{agent})
	defer plane.Close()

	runtimeID := RuntimeIDForAgent("daemon-1", agent)
	req, err := plane.ClaimTask(context.Background(), runtimeID)
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      runtimeID,
		RunningTaskIDs: []string{req.ID},
	}); err != nil {
		t.Fatalf("Heartbeat should retry transient heartbeat: %v", err)
	}
	if got := fake.requestCount("/v1/agents/jykim1/heartbeat"); got != 2 {
		t.Fatalf("heartbeat request count = %d, want 2", got)
	}
}

func TestPlaneRetriesTransientRuntimeSnapshot(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	fake.failNext("/v1/daemon/runtime-snapshot", 1, http.StatusTooManyRequests)
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

	err = plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "daemon-1:codex",
		Provider:  "codex",
	})
	if err != nil {
		t.Fatalf("RegisterRuntime should retry transient runtime snapshot: %v", err)
	}
	if got := fake.requestCount("/v1/daemon/runtime-snapshot"); got != 2 {
		t.Fatalf("runtime snapshot request count = %d, want 2", got)
	}
}

func TestPlaneDoesNotRetryPermanentPollFailure(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.failNext("/v1/agents/jykim1/poll", 1, http.StatusUnauthorized)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	_, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("ClaimTask should return permanent auth failure without retry, got %v", err)
	}
	if got := fake.requestCount("/v1/agents/jykim1/poll"); got != 1 {
		t.Fatalf("poll request count = %d, want 1", got)
	}
}

func TestPlaneSendsLongPollWaitMsAndExtendsRequestTimeout(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane, err := New(Config{
		BaseURL:        fake.URL(),
		DaemonID:       "daemon-1",
		DeviceID:       "device-1",
		Agents:         []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}},
		RequestTimeout: time.Second,
		LongPollWait:   2500 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req != nil {
		t.Fatalf("empty fake queue should not claim task: %+v", req)
	}
	polls := fake.pollRequestsFor("jykim1")
	if len(polls) != 1 {
		t.Fatalf("poll requests = %+v", polls)
	}
	if polls[0].WaitMs != 2500 {
		t.Fatalf("wait_ms = %d, want 2500", polls[0].WaitMs)
	}
	if plane.cfg.RequestTimeout != 7500*time.Millisecond {
		t.Fatalf("request timeout = %s, want 7.5s", plane.cfg.RequestTimeout)
	}
}

func TestPlaneLongPollsStaticCandidatesConcurrently(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane, err := New(Config{
		BaseURL:  fake.URL(),
		DaemonID: "daemon-1",
		DeviceID: "device-1",
		Agents: []AgentBinding{
			{AgentID: "agent-a", RuntimeProvider: "codex"},
			{AgentID: "agent-b", RuntimeProvider: "codex"},
		},
		LongPollWait: 2500 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req != nil {
		t.Fatalf("empty fake queue should not claim task: %+v", req)
	}
	agentA := fake.pollRequestsFor("agent-a")
	agentB := fake.pollRequestsFor("agent-b")
	if len(agentA) != 1 || len(agentB) != 1 {
		t.Fatalf("poll requests agent-a=%+v agent-b=%+v", agentA, agentB)
	}
	if agentA[0].WaitMs != 2500 || agentB[0].WaitMs != 2500 {
		t.Fatalf("unexpected wait_ms distribution agent-a=%+v agent-b=%+v", agentA, agentB)
	}
}

func TestPlaneDoesNotRetryEventPostWithoutIdempotency(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "hello",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	fake.failNext("/v1/agents/jykim1/events", 1, http.StatusServiceUnavailable)
	err = plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{Kind: agentbridge.EventProgress, Text: "progress"})
	if err == nil || !strings.Contains(err.Error(), "503") {
		t.Fatalf("ReportEvent should return first transient event failure without retry, got %v", err)
	}
	if got := fake.requestCount("/v1/agents/jykim1/events"); got != 1 {
		t.Fatalf("event request count = %d, want 1", got)
	}
}

func TestPlaneRegistersRuntimeSnapshotWithDeviceCredential(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	startedAt := time.Now().Add(-2 * time.Minute).UTC()
	plane, err := New(Config{
		BaseURL:      fake.URL(),
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
		Profile:      "development",
		AppVersion:   "v0.0.13",
		PID:          4321,
		StartedAt:    startedAt,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	err = plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
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
	})
	if err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}
	if len(fake.runtimeSnapshots) != 1 {
		t.Fatalf("runtime snapshots = %+v", fake.runtimeSnapshots)
	}
	snapshot := fake.runtimeSnapshots[0]
	if snapshot.DaemonID != "daemon-1" || snapshot.DeviceID != "device-1" || snapshot.DeviceDisplayName != "주윤의 MacBook" {
		t.Fatalf("snapshot identity = %+v", snapshot)
	}
	if snapshot.Profile != "development" || snapshot.AppVersion != "v0.0.13" || snapshot.PID != 4321 || !snapshot.StartedAt.Equal(startedAt) || snapshot.UptimeSeconds <= 0 {
		t.Fatalf("snapshot daemon facts = %+v", snapshot)
	}
	if len(snapshot.Runtimes) != 1 ||
		snapshot.Runtimes[0].RuntimeID != "daemon-1:codex" ||
		snapshot.Runtimes[0].Kind != "codex" ||
		snapshot.Runtimes[0].Availability != "online" ||
		snapshot.Runtimes[0].DetectionState != "detected" ||
		snapshot.Runtimes[0].ProviderVersion != "codex-cli 0.133.0" ||
		!snapshot.Runtimes[0].RequiresExperimentalOptIn {
		t.Fatalf("snapshot runtimes = %+v", snapshot.Runtimes)
	}
	if len(snapshot.Runtimes[0].Models) != 1 || snapshot.Runtimes[0].Models[0].ModelID != "gpt-5.5" || !snapshot.Runtimes[0].Models[0].IsDefault {
		t.Fatalf("snapshot runtime models = %+v", snapshot.Runtimes[0].Models)
	}
}
