package saasplane

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestPlaneSkipsStaleDynamicCandidateAndClaimsLiveAgent(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{
		codexRuntimeBinding("agent-stale-codex"),
		codexRuntimeBinding("agent-live-codex"),
	}
	fake.failNext("/v1/agents/agent-stale-codex/poll", 1, http.StatusBadRequest)
	fake.replaceBindingsAfterNextFailure(
		"/v1/agents/agent-stale-codex/poll",
		[]assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("agent-live-codex")},
	)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-live",
		TaskID:          "task-live",
		ComponentID:     "component-live",
		AgentID:         "agent-live-codex",
		RuntimeProvider: "codex",
		Prompt:          "use provider document context",
		State:           assignmentcontract.AssignmentQueued,
	})
	plane := newRuntimeBindingPlane(t, fake, nil)

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req == nil || req.ID != "asn-live" {
		t.Fatalf("ClaimTask request = %+v, want asn-live", req)
	}
	if got := len(fake.pollRequestsFor("agent-live-codex")); got == 0 {
		t.Fatal("live candidate was not polled after stale binding rejection")
	}
	if got := len(fake.pollRequestsFor("agent-stale-codex")); got != 0 {
		t.Fatalf("stale candidate successful poll count = %d, want 0", got)
	}
	if got := fake.requestCount("/v1/agents/agent-stale-codex/poll"); got != 1 {
		t.Fatalf("stale candidate request count = %d, want 1", got)
	}
	if got := fake.requestCount("/v1/daemon/agent-bindings"); got != 2 {
		t.Fatalf("agent-bindings request count = %d, want initial plus reconciliation", got)
	}
}

func TestPlaneSuppressesRejectedDynamicCandidateUntilBindingCacheRefresh(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("agent-stale-codex")}
	fake.failNext("/v1/agents/agent-stale-codex/poll", 2, http.StatusBadRequest)
	plane := newRuntimeBindingPlane(t, fake, nil)

	for i := range 2 {
		req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
		if err != nil {
			t.Fatalf("ClaimTask %d: %v", i+1, err)
		}
		if req != nil {
			t.Fatalf("ClaimTask %d request = %+v, want nil", i+1, req)
		}
	}
	if got := fake.requestCount("/v1/agents/agent-stale-codex/poll"); got != 1 {
		t.Fatalf("stale candidate request count = %d, want one bounded rejection", got)
	}
	if got := fake.requestCount("/v1/daemon/agent-bindings"); got != 2 {
		t.Fatalf("agent-bindings request count = %d, want initial plus reconciliation", got)
	}
}

func TestPlaneSkipsMultipleStaleCandidatesAndClaimsLiveAgent(t *testing.T) {
	for _, mode := range []string{"dynamic unchanged", "dynamic removed", "static"} {
		t.Run(mode, func(t *testing.T) {
			// Exercise the HTTP handlers in memory so this regression needs no listener.
			fake := &fakeAssignmentServer{
				t:                    t,
				assignmentsByAgent:   map[string][]assignmentcontract.Assignment{},
				assignmentsByID:      map[string]assignmentcontract.Assignment{},
				requestCounts:        map[string]int{},
				transientFailures:    map[string]int{},
				transientStatuses:    map[string]int{},
				bindingsAfterFailure: map[string][]assignmentcontract.AgentRuntimeBinding{},
				pollRequestsByAgent:  map[string][]assignmentcontract.PollRequest{},
			}
			fake.bindings = []assignmentcontract.AgentRuntimeBinding{
				codexRuntimeBinding("agent-stale-a"),
				codexRuntimeBinding("agent-stale-b"),
				codexRuntimeBinding("agent-live-codex"),
			}
			for _, agentID := range []string{"agent-stale-a", "agent-stale-b"} {
				fake.failNext("/v1/agents/"+agentID+"/poll", 2, http.StatusBadRequest)
			}
			if mode == "dynamic removed" {
				fake.replaceBindingsAfterNextFailure("/v1/agents/agent-stale-a/poll", fake.bindings[1:])
			}
			fake.enqueue(assignmentcontract.Assignment{
				ID:              "asn-live",
				TaskID:          "task-live",
				ComponentID:     "component-live",
				AgentID:         "agent-live-codex",
				RuntimeProvider: "codex",
				Prompt:          "use provider document context",
				State:           assignmentcontract.AssignmentQueued,
			})
			cfg := Config{
				BaseURL:      "http://saasplane.invalid",
				DaemonID:     "daemon-1",
				DeviceID:     "device-1",
				DeviceSecret: "synthetic-device-secret",
				HTTPClient:   &http.Client{Transport: staleCandidateHandlerTransport{http.HandlerFunc(fake.handle)}},
			}
			if mode == "static" {
				cfg.Agents = []AgentBinding{
					{AgentID: "agent-stale-a", RuntimeProvider: "codex"},
					{AgentID: "agent-stale-b", RuntimeProvider: "codex"},
					{AgentID: "agent-live-codex", RuntimeProvider: "codex"},
				}
			}
			plane, err := New(cfg)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			t.Cleanup(plane.Close)

			req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
			if err != nil {
				t.Fatalf("ClaimTask: %v", err)
			}
			if req == nil || req.ID != "asn-live" {
				t.Fatalf("ClaimTask request = %+v, want asn-live", req)
			}
			for _, agentID := range []string{"agent-stale-a", "agent-stale-b", "agent-live-codex"} {
				if got := fake.requestCount("/v1/agents/" + agentID + "/poll"); got != 1 {
					t.Fatalf("%s poll count = %d, want 1", agentID, got)
				}
			}
			wantBindingRequests := 2
			if mode == "static" {
				wantBindingRequests = 0
			}
			if got := fake.requestCount("/v1/daemon/agent-bindings"); got != wantBindingRequests {
				t.Fatalf("agent-bindings request count = %d, want %d", got, wantBindingRequests)
			}
		})
	}
}

type staleCandidateHandlerTransport struct {
	handler http.Handler
}

func (t staleCandidateHandlerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		defer req.Body.Close()
	}
	recorder := httptest.NewRecorder()
	t.handler.ServeHTTP(recorder, req)
	return recorder.Result(), nil
}
