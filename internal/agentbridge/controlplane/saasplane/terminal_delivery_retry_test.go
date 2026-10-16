package saasplane

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// The server commits each request, but its acknowledgements are lost beyond
// the HTTP retry budget. The next supervisor attempt must replay the same key.
func TestTerminalDeliveryKeepsKeyAfterExhaustedHTTPRetries(t *testing.T) {
	fake, plane, req := newTextDeltaFixture(t)
	plane.client = &http.Client{Transport: &lostTerminalAcknowledgements{remaining: 3}}
	result := agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "Finished answer"}
	if err := plane.CompleteTask(context.Background(), req.ID, result); err == nil {
		t.Fatal("expected exhausted transport retries")
	}
	if err := plane.CompleteTask(context.Background(), req.ID, result); err != nil {
		t.Fatal(err)
	}
	if len(fake.events) != 4 {
		t.Fatalf("events=%d", len(fake.events))
	}
	key := fake.events[0].Metadata[metadatakeys.AssignmentEventKey.String()]
	if !strings.HasPrefix(key, "terminal:") {
		t.Fatal("missing stable terminal key")
	}
	for _, event := range fake.events {
		if event.Metadata[metadatakeys.AssignmentEventKey.String()] != key {
			t.Fatal("supervisor retry changed idempotency key")
		}
	}
}

type lostTerminalAcknowledgements struct{ remaining int }

func (t *lostTerminalAcknowledgements) RoundTrip(r *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(r)
	if err != nil || t.remaining == 0 {
		return response, err
	}
	t.remaining--
	response.Body.Close()
	return &http.Response{StatusCode: http.StatusServiceUnavailable, Status: "503 Service Unavailable", Header: make(http.Header), Body: io.NopCloser(strings.NewReader("committed; projection pending")), Request: r}, nil
}

func TestTerminalDeliveryKeySeparatesLeaseAndOutcome(t *testing.T) {
	assignment := textDeltaAssignment()
	request := assignmentcontract.AgentEventRequest{AssignmentID: assignment.ID, TaskID: assignment.TaskID, State: assignmentcontract.AssignmentCompleted, Message: "answer"}
	key := terminalEventKey(assignment, "execution", request)
	assignment.LeaseToken = "new-lease"
	if key == terminalEventKey(assignment, "execution", request) {
		t.Fatal("lease ownership collapsed")
	}
	assignment.LeaseToken = "lease-1"
	request.Message = "different result"
	if key == terminalEventKey(assignment, "execution", request) {
		t.Fatal("different outcomes collapsed")
	}
}
