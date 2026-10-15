package saasplane

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func readyAcknowledgementFixture() (assignmentcontract.Assignment, assignmentcontract.AgentEventRequest, assignmentcontract.AgentEventResponse) {
	a := saasClaimReportAssignment()
	a.State = assignmentcontract.AssignmentReady
	req := assignmentcontract.AgentEventRequest{Metadata: map[string]string{"assignment_event_key": readyEventKey(a, assignmentExecutionID(a))}}
	response := assignmentcontract.AgentEventResponse{SchemaVersion: assignmentcontract.SchemaVersion, Assignment: &a, Event: assignmentcontract.TaskEvent{EventID: a.ID + ":1", AssignmentID: a.ID, TaskID: a.TaskID, AgentID: a.AgentID, Type: assignmentcontract.EventAssignmentReady, State: assignmentcontract.AssignmentReady, Seq: 1, At: time.Now().UTC(), Metadata: cloneEventMetadata(req.Metadata)}}
	return a, req, response
}

func TestReadyAcknowledgementRejectsEmptyAndForeignAuthority(t *testing.T) {
	a, req, response := readyAcknowledgementFixture()
	if err := validateReadyAcknowledgement(a, req, response); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(*assignmentcontract.AgentEventResponse){
		"empty":      func(r *assignmentcontract.AgentEventResponse) { *r = assignmentcontract.AgentEventResponse{} },
		"schema":     func(r *assignmentcontract.AgentEventResponse) { r.SchemaVersion = "other" },
		"assignment": func(r *assignmentcontract.AgentEventResponse) { r.Event.AssignmentID = "foreign" },
		"task":       func(r *assignmentcontract.AgentEventResponse) { r.Event.TaskID = "foreign" },
		"agent":      func(r *assignmentcontract.AgentEventResponse) { r.Event.AgentID = "foreign" },
		"event": func(r *assignmentcontract.AgentEventResponse) {
			r.Event.Type = assignmentcontract.EventAssignmentRunning
		},
		"state":            func(r *assignmentcontract.AgentEventResponse) { r.Event.State = assignmentcontract.AssignmentRunning },
		"sequence":         func(r *assignmentcontract.AgentEventResponse) { r.Event.Seq = 0 },
		"missing_event_id": func(r *assignmentcontract.AgentEventResponse) { r.Event.EventID = "" },
		"foreign_event_id": func(r *assignmentcontract.AgentEventResponse) { r.Event.EventID = "foreign:1" },
		"unsanitized_lease": func(r *assignmentcontract.AgentEventResponse) {
			r.Event.Metadata = cloneEventMetadata(r.Event.Metadata)
			r.Event.Metadata["lease_token"] = "secret"
		},
		"time": func(r *assignmentcontract.AgentEventResponse) { r.Event.At = time.Time{} },
		"key": func(r *assignmentcontract.AgentEventResponse) {
			r.Event.Metadata = map[string]string{"assignment_event_key": "foreign"}
		},
		"missing_authority": func(r *assignmentcontract.AgentEventResponse) { r.Assignment = nil },
		"lease": func(r *assignmentcontract.AgentEventResponse) {
			copyAssignment := *r.Assignment
			copyAssignment.LeaseToken = "other"
			r.Assignment = &copyAssignment
		},
	}
	for _, field := range []string{"runtime_provider", "model_id", "request_id", "component_id", "workspace_id", "team_id", "created_by"} {
		cases["assignment_scope_"+field] = func(r *assignmentcontract.AgentEventResponse) {
			cloned := *r.Assignment
			switch field {
			case "runtime_provider":
				cloned.RuntimeProvider = "foreign"
			case "model_id":
				cloned.ModelID = "foreign"
			case "request_id":
				cloned.RequestID = "foreign"
			case "component_id":
				cloned.ComponentID = "foreign"
			case "workspace_id":
				cloned.WorkspaceID = "foreign"
			case "team_id":
				cloned.TeamID = "foreign"
			case "created_by":
				cloned.CreatedBy = "foreign"
			}
			r.Assignment = &cloned
		}
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, changed := readyAcknowledgementFixture()
			mutate(&changed)
			if err := validateReadyAcknowledgement(a, req, changed); err == nil {
				t.Fatal("unsafe Ready ACK accepted")
			}
		})
	}
}

func TestReadyRetriesKeepSamePayloadAndLeaseScopedKeyAfterHTTPExhaustion(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	a := saasClaimReportAssignment()
	a.ResumeSessionID = ""
	a.Worktree = nil
	fake.enqueue(a)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: a.AgentID, RuntimeProvider: "codex"}})
	defer plane.Close()
	request, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil || request == nil {
		t.Fatal("claim failed", err)
	}
	path := "/v1/agents/" + a.AgentID + "/events"
	fake.failNext(path, 3, http.StatusServiceUnavailable)
	err = plane.StartTask(context.Background(), request.ID)
	var retryable interface{ RetryableReady() bool }
	if !errors.As(err, &retryable) || !retryable.RetryableReady() {
		t.Fatalf("exhausted503 not retryable: %v", err)
	}
	if err = plane.StartTask(context.Background(), request.ID); err != nil {
		t.Fatal(err)
	}
	if err = plane.StartTask(context.Background(), request.ID); err != nil {
		t.Fatal(err)
	}
	fake.mu.Lock()
	events := append([]assignmentcontract.AgentEventRequest(nil), fake.events...)
	fake.mu.Unlock()
	if len(events) != 2 || events[0].Metadata["assignment_event_key"] != events[1].Metadata["assignment_event_key"] || events[0].Metadata["lease_token"] != a.LeaseToken {
		t.Fatal("Ready retries changed stable authority payload")
	}
	if fake.requestCount(path) != 5 {
		t.Fatalf("requests=%d", fake.requestCount(path))
	}
	newer := a
	newer.LeaseToken = "new-lease"
	if readyEventKey(a, request.ID) == readyEventKey(newer, request.ID) {
		t.Fatal("Ready key crossed lease boundary")
	}
	a.State = assignmentcontract.AssignmentReady
	fake.activeNext(a.AgentID, a)
	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil || req != nil {
		t.Fatal("stateless no-session Active replayed", err)
	}
}

func TestReadyMissingLocalAuthorityIsPermanent(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()
	err := plane.StartTask(context.Background(), "unknown")
	var retryable interface{ RetryableReady() bool }
	if !errors.As(err, &retryable) || retryable.RetryableReady() {
		t.Fatalf("authority failure=%v", err)
	}
}

func TestReadyCommittedButAllHTTPAcknowledgementsLostRetriesIdenticalEvent(t *testing.T) {
	a, _, ack := readyAcknowledgementFixture()
	var mu sync.Mutex
	var requests []assignmentcontract.AgentEventRequest
	commits := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req assignmentcontract.AgentEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid fixture", http.StatusBadRequest)
			return
		}
		mu.Lock()
		requests = append(requests, req)
		commits[req.Metadata["assignment_event_key"]] = true
		attempt := len(requests)
		mu.Unlock()
		if attempt <= 3 {
			connection, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			_ = connection.Close()
			return
		}
		response := ack
		response.Event.Metadata = map[string]string{"assignment_event_key": req.Metadata["assignment_event_key"]}
		writeJSON(w, response)
	}))
	defer server.Close()
	plane := newTestPlane(t, server.URL, []AgentBinding{{AgentID: a.AgentID, RuntimeProvider: "codex"}})
	defer plane.Close()
	executionID := assignmentExecutionID(a)
	if err := plane.saveAssignmentRuntime(context.Background(), a, "daemon-1:codex"); err != nil {
		t.Fatal(err)
	}
	err := plane.StartTask(context.Background(), executionID)
	var retryable interface{ RetryableReady() bool }
	if !errors.As(err, &retryable) || !retryable.RetryableReady() {
		t.Fatalf("lost committed ACK=%v", err)
	}
	if err = plane.StartTask(context.Background(), executionID); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 4 || len(commits) != 1 {
		t.Fatalf("requests=%d committed semantic events=%d", len(requests), len(commits))
	}
	for _, req := range requests[1:] {
		if !reflect.DeepEqual(req, requests[0]) {
			t.Fatal("acknowledgement retry changed Ready payload")
		}
	}
}
