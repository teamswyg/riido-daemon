package saasplane

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRecoveryAcknowledgementRejectsMalformedOrForeignSuccess(t *testing.T) {
	assignment := saasClaimReportAssignment()
	c := assignmentcontract.RecoveryContext{State: "checking", Reason: "none", Provider: "codex", Circuit: "closed", OutcomeCertainty: "not_started", Attempt: 1}
	expected := c.Metadata()
	expected["assignment_event_key"] = "expected-key"
	request := assignmentcontract.AgentEventRequest{Metadata: expected}
	valid := func() assignmentcontract.AgentEventResponse {
		return assignmentcontract.AgentEventResponse{SchemaVersion: assignmentcontract.SchemaVersion, Event: assignmentcontract.TaskEvent{AssignmentID: assignment.ID, TaskID: assignment.TaskID, AgentID: assignment.AgentID, Type: assignmentcontract.EventRiidoLog, Seq: 1, At: time.Now(), Metadata: cloneEventMetadata(expected)}}
	}
	if err := validateRecoveryAcknowledgement(assignment, request, valid()); err != nil {
		t.Fatal(err)
	}
	mutations := map[string]func(*assignmentcontract.AgentEventResponse){
		"empty":      func(r *assignmentcontract.AgentEventResponse) { *r = assignmentcontract.AgentEventResponse{} },
		"schema":     func(r *assignmentcontract.AgentEventResponse) { r.SchemaVersion = "unknown" },
		"assignment": func(r *assignmentcontract.AgentEventResponse) { r.Event.AssignmentID = "another" },
		"task":       func(r *assignmentcontract.AgentEventResponse) { r.Event.TaskID = "another" },
		"agent":      func(r *assignmentcontract.AgentEventResponse) { r.Event.AgentID = "another" },
		"event":      func(r *assignmentcontract.AgentEventResponse) { r.Event.Type = assignmentcontract.EventProviderLog },
		"key":        func(r *assignmentcontract.AgentEventResponse) { r.Event.Metadata["assignment_event_key"] = "another" },
		"context": func(r *assignmentcontract.AgentEventResponse) {
			r.Event.Metadata[assignmentcontract.RecoveryAttemptKey] = "2"
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			response := valid()
			mutate(&response)
			if err := validateRecoveryAcknowledgement(assignment, request, response); err == nil {
				t.Fatal("foreign acknowledgement accepted")
			}
		})
	}
}

func TestRecoveryHTTP200WithoutAcknowledgementFailsClosed(t *testing.T) {
	for _, payload := range []string{"{}", `{"schema_version":"riido.assignment.v1","event":{"assignment_id":"foreign"}}`} {
		t.Run(payload, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				var value any
				if err := json.Unmarshal([]byte(payload), &value); err != nil {
					t.Error(err)
				}
				_ = json.NewEncoder(w).Encode(value)
			}))
			defer server.Close()
			plane := newTestPlane(t, server.URL, []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
			defer plane.Close()
			assignment := saasClaimReportAssignment()
			id := assignmentExecutionID(assignment)
			if err := plane.withState(context.Background(), func(s *planeState) { s.assignmentsByExecution[id] = assignment }); err != nil {
				t.Fatal(err)
			}
			c := assignmentcontract.RecoveryContext{State: "checking", Reason: "none", Provider: "codex", Circuit: "closed", OutcomeCertainty: "not_started", Attempt: 1}
			if err := plane.ReportEvent(context.Background(), id, agentbridge.Event{Kind: agentbridge.EventProgress, RecoveryContext: &c}); err == nil {
				t.Fatal("HTTP success without a matching stored journal authorized dispatch")
			}
		})
	}
}
