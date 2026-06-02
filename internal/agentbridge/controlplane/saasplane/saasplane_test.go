package saasplane

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneClaimsAndReportsAssignment(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:               "asn-1",
		TaskID:           "task-a",
		ComponentID:      "component-1",
		AgentID:          "jykim1",
		RuntimeProvider:  "codex",
		Prompt:           "golang hello world quickly",
		AgentInstruction: "write concise Korean progress updates",
		State:            assignmentcontract.AssignmentQueued,
		LeaseToken:       "lease-1",
	})
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req == nil || req.ID != "task-a" || req.Provider != "codex" {
		t.Fatalf("request = %+v", req)
	}
	if got := req.Metadata[MetadataAssignmentID]; got != "asn-1" {
		t.Fatalf("assignment metadata = %q", got)
	}
	if got := req.Metadata["workspace_id"]; got != "component-1" {
		t.Fatalf("workspace_id = %q", got)
	}
	if !strings.Contains(req.Prompt, "<riido_log>") || !strings.Contains(req.Prompt, "golang hello world") || !strings.Contains(req.Prompt, "write concise Korean progress updates") {
		t.Fatalf("prompt missing telemetry contract: %q", req.Prompt)
	}
	if got := req.Metadata[agentbridge.MetadataTelemetryContract]; got != agentbridge.TelemetryPlacementPrompt {
		t.Fatalf("telemetry placement = %q", got)
	}
	if got := req.Metadata[agentbridge.MetadataAgentInstruction]; got != agentbridge.TelemetryPlacementPrompt {
		t.Fatalf("instruction placement = %q", got)
	}

	if err := plane.StartTask(context.Background(), req.ID); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      RuntimeIDForAgent("daemon-1", AgentBinding{AgentID: "jykim1", RuntimeProvider: "codex"}),
		RunningTaskIDs: []string{req.ID},
	}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	if err := plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{Kind: agentbridge.EventProgress, Text: "project go.mod written"}); err != nil {
		t.Fatalf("ReportEvent: %v", err)
	}
	if err := plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}); err != nil {
		t.Fatalf("ReportEvent running: %v", err)
	}
	if err := plane.CompleteTask(context.Background(), req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "ok"}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	fake.assertEvent(t, assignmentcontract.EventAssignmentReady)
	fake.assertEvent(t, assignmentcontract.EventRiidoLog)
	fake.assertEvent(t, assignmentcontract.EventAssignmentRunning)
	fake.assertEvent(t, assignmentcontract.EventAssignmentCompleted)
	heartbeats := fake.heartbeatsFor("jykim1")
	if len(heartbeats) != 1 || len(heartbeats[0].ActiveAssignmentIDs) != 1 || heartbeats[0].ActiveAssignmentIDs[0] != "asn-1" {
		t.Fatalf("heartbeats = %+v", heartbeats)
	}
}

func TestTaskRequestPlacesTelemetryForSystemPromptProviders(t *testing.T) {
	assignment := assignmentcontract.Assignment{
		ID:               "asn-1",
		TaskID:           "task-a",
		ComponentID:      "component-1",
		AgentID:          "jykim1",
		RuntimeProvider:  "claude",
		Prompt:           "golang hello world quickly",
		AgentInstruction: "act as a backend reviewer",
	}
	req := taskRequestFromAssignment(assignment)
	if req.Prompt != assignment.Prompt {
		t.Fatalf("claude prompt should remain user task only: %q", req.Prompt)
	}
	if !strings.Contains(req.SystemPrompt, "<riido_log>") || !strings.Contains(req.SystemPrompt, "act as a backend reviewer") {
		t.Fatalf("claude system prompt missing runtime instructions: %q", req.SystemPrompt)
	}
	if got := req.Metadata[agentbridge.MetadataTelemetryContract]; got != agentbridge.TelemetryPlacementSystemPrompt {
		t.Fatalf("telemetry placement = %q", got)
	}
	if got := req.Metadata[agentbridge.MetadataAgentInstruction]; got != agentbridge.TelemetryPlacementSystemPrompt {
		t.Fatalf("instruction placement = %q", got)
	}
}

func TestPlaneDeliversCancellationFromPollResponse(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	first := assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "first",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	}
	fake.enqueue(first)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask first: %v", err)
	}
	if req == nil || req.Metadata[MetadataAssignmentID] != first.ID {
		t.Fatalf("first claim = %+v", req)
	}
	if err := plane.StartTask(context.Background(), req.ID); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	cancelCh, err := plane.WatchCancellation(context.Background(), req.ID)
	if err != nil {
		t.Fatalf("WatchCancellation: %v", err)
	}

	fake.cancelNext(first.AgentID, first)
	req, err = plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask cancel poll: %v", err)
	}
	if req != nil {
		t.Fatalf("cancel poll should not claim new task: %+v", req)
	}
	select {
	case cause := <-cancelCh:
		if cause == nil || !strings.Contains(cause.Error(), first.ID) {
			t.Fatalf("cancel cause = %v", cause)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for cancellation")
	}
}

func TestPlanePollsOnlyRuntimeScopedAgent(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-2",
		TaskID:          "task-b",
		ComponentID:     "component-1",
		AgentID:         "jykim2",
		RuntimeProvider: "codex",
		Prompt:          "second agent task",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-2",
	})
	agents := []AgentBinding{
		{AgentID: "jykim1", RuntimeProvider: "codex"},
		{AgentID: "jykim2", RuntimeProvider: "codex"},
	}
	plane := newTestPlane(t, fake.URL(), agents)
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), RuntimeIDForAgent("daemon-1", agents[0]))
	if err != nil {
		t.Fatalf("ClaimTask jykim1: %v", err)
	}
	if req != nil {
		t.Fatalf("jykim1 runtime claimed another agent task: %+v", req)
	}
	req, err = plane.ClaimTask(context.Background(), RuntimeIDForAgent("daemon-1", agents[1]))
	if err != nil {
		t.Fatalf("ClaimTask jykim2: %v", err)
	}
	if req == nil || req.ID != "task-b" || req.Metadata[MetadataAgentID] != "jykim2" {
		t.Fatalf("jykim2 claim = %+v", req)
	}
}

func TestPlaneSendsBearerToken(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bearerToken = "secret"
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
	plane := newTestPlaneWithToken(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}}, "secret")
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask with token: %v", err)
	}
	if req == nil || req.ID != "task-a" {
		t.Fatalf("request = %+v", req)
	}
}

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
	if req == nil || req.ID != "task-a" {
		t.Fatalf("request = %+v", req)
	}
}

func TestPlaneRegistersRuntimeSnapshotWithDeviceCredential(t *testing.T) {
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

	err = plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		DaemonID:   "daemon-1",
		RuntimeID:  "daemon-1:codex",
		Provider:   "codex",
		DeviceName: "주윤의 MacBook",
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
	if len(snapshot.Runtimes) != 1 || snapshot.Runtimes[0].RuntimeID != "daemon-1:codex" || snapshot.Runtimes[0].Kind != "codex" {
		t.Fatalf("snapshot runtimes = %+v", snapshot.Runtimes)
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
	if req == nil || req.ID != "task-a" || req.Provider != "codex" || req.Metadata[MetadataAgentID] != "jykim1" {
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

func newTestPlaneWithToken(t *testing.T, baseURL string, agents []AgentBinding, token string) *Plane {
	t.Helper()
	plane, err := New(Config{
		BaseURL:     baseURL,
		DaemonID:    "daemon-1",
		DeviceID:    "device-1",
		Agents:      agents,
		BearerToken: token,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return plane
}

type fakeAssignmentServer struct {
	t            *testing.T
	server       *httptest.Server
	bearerToken  string
	deviceID     string
	deviceSecret string

	assignmentsByAgent map[string][]assignmentcontract.Assignment
	cancelByAgent      map[string]assignmentcontract.Assignment
	bindings           []assignmentcontract.AgentRuntimeBinding
	runtimeSnapshots   []DeviceRuntimeSnapshotSyncRequest
	events             []assignmentcontract.AgentEventRequest
	heartbeats         []assignmentcontract.AgentHeartbeatRequest
}

func newFakeAssignmentServer(t *testing.T) *fakeAssignmentServer {
	t.Helper()
	f := &fakeAssignmentServer{
		t:                  t,
		assignmentsByAgent: map[string][]assignmentcontract.Assignment{},
		cancelByAgent:      map[string]assignmentcontract.Assignment{},
	}
	f.server = httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeAssignmentServer) URL() string {
	return f.server.URL
}

func (f *fakeAssignmentServer) enqueue(assignment assignmentcontract.Assignment) {
	f.assignmentsByAgent[assignment.AgentID] = append(f.assignmentsByAgent[assignment.AgentID], assignment)
}

func (f *fakeAssignmentServer) cancelNext(agentID string, assignment assignmentcontract.Assignment) {
	assignment.State = assignmentcontract.AssignmentCancelling
	f.cancelByAgent[agentID] = assignment
}

func (f *fakeAssignmentServer) handle(w http.ResponseWriter, r *http.Request) {
	if f.deviceSecret != "" && (r.Header.Get("X-Riido-Device-ID") != f.deviceID || r.Header.Get("X-Riido-Device-Secret") != f.deviceSecret) {
		http.Error(w, "missing device credential", http.StatusUnauthorized)
		return
	}
	if f.bearerToken != "" && r.Header.Get("Authorization") != "Bearer "+f.bearerToken {
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
		return
	}
	if strings.Trim(r.URL.Path, "/") == "v1/daemon/agent-bindings" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, AgentRuntimeBindingListResponse{
			SchemaVersion: assignmentcontract.SchemaVersion,
			Bindings:      append([]assignmentcontract.AgentRuntimeBinding(nil), f.bindings...),
		})
		return
	}
	if strings.Trim(r.URL.Path, "/") == "v1/daemon/runtime-snapshot" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req DeviceRuntimeSnapshotSyncRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.runtimeSnapshots = append(f.runtimeSnapshots, req)
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, struct {
			SchemaVersion string `json:"schema_version"`
		}{SchemaVersion: assignmentcontract.SchemaVersion})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "v1" || parts[1] != "agents" {
		http.NotFound(w, r)
		return
	}
	agentID, err := url.PathUnescape(parts[2])
	if err != nil {
		http.Error(w, "bad agent id", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	switch parts[3] {
	case "poll":
		f.handlePoll(w, r, agentID)
	case "heartbeat":
		f.handleHeartbeat(w, r)
	case "events":
		f.handleEvents(w, r, agentID)
	default:
		http.NotFound(w, r)
	}
}

func (f *fakeAssignmentServer) handlePoll(w http.ResponseWriter, r *http.Request, agentID string) {
	var req assignmentcontract.PollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.DaemonID == "" || req.RuntimeID == "" {
		http.Error(w, "missing poll identity", http.StatusBadRequest)
		return
	}
	if cancel, ok := f.cancelByAgent[agentID]; ok {
		delete(f.cancelByAgent, agentID)
		writeJSON(w, assignmentcontract.PollResponse{
			SchemaVersion: assignmentcontract.SchemaVersion,
			Action:        assignmentcontract.PollCancel,
			Assignment:    &cancel,
		})
		return
	}
	queue := f.assignmentsByAgent[agentID]
	if len(queue) == 0 {
		writeJSON(w, assignmentcontract.PollResponse{
			SchemaVersion: assignmentcontract.SchemaVersion,
			Action:        assignmentcontract.PollNone,
		})
		return
	}
	assignment := queue[0]
	f.assignmentsByAgent[agentID] = queue[1:]
	assignment.State = assignmentcontract.AssignmentLeased
	writeJSON(w, assignmentcontract.PollResponse{
		SchemaVersion: assignmentcontract.SchemaVersion,
		Action:        assignmentcontract.PollStart,
		Assignment:    &assignment,
	})
}

func (f *fakeAssignmentServer) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req assignmentcontract.AgentHeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	f.heartbeats = append(f.heartbeats, req)
	writeJSON(w, assignmentcontract.AgentHeartbeatResponse{
		SchemaVersion: assignmentcontract.SchemaVersion,
	})
}

func (f *fakeAssignmentServer) handleEvents(w http.ResponseWriter, r *http.Request, agentID string) {
	var req assignmentcontract.AgentEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.AssignmentID == "" || req.TaskID == "" || req.DaemonID == "" || req.RuntimeID == "" {
		http.Error(w, "missing event identity", http.StatusBadRequest)
		return
	}
	f.events = append(f.events, req)
	writeJSON(w, assignmentcontract.AgentEventResponse{
		SchemaVersion: assignmentcontract.SchemaVersion,
		Event: assignmentcontract.TaskEvent{
			Seq:          int64(len(f.events)),
			TaskID:       req.TaskID,
			AssignmentID: req.AssignmentID,
			AgentID:      agentID,
			Type:         req.EventType,
			State:        req.State,
			Message:      req.Message,
			At:           time.Now().UTC(),
		},
	})
}

func (f *fakeAssignmentServer) assertEvent(t *testing.T, eventType string) {
	t.Helper()
	for _, ev := range f.events {
		if ev.EventType == eventType {
			return
		}
	}
	t.Fatalf("event %q missing from %+v", eventType, f.events)
}

func (f *fakeAssignmentServer) heartbeatsFor(agentID string) []assignmentcontract.AgentHeartbeatRequest {
	var out []assignmentcontract.AgentHeartbeatRequest
	for _, hb := range f.heartbeats {
		if runtimeAgent, ok := agentFromRuntimeID(hb.RuntimeID); ok && runtimeAgent == agentID {
			out = append(out, hb)
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, value any) {
	_ = json.NewEncoder(w).Encode(value)
}
