package saasplane

import (
	"net/http"
	"net/http/httptest"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func newFakeAssignmentServer(t *testing.T) *fakeAssignmentServer {
	t.Helper()
	f := &fakeAssignmentServer{
		t:                   t,
		assignmentsByAgent:  map[string][]assignmentcontract.Assignment{},
		assignmentsByID:     map[string]assignmentcontract.Assignment{},
		activeByAgent:       map[string]assignmentcontract.Assignment{},
		cancelByAgent:       map[string]assignmentcontract.Assignment{},
		staleHeartbeatIDs:   map[string]bool{},
		requestCounts:       map[string]int{},
		transientFailures:   map[string]int{},
		transientStatuses:   map[string]int{},
		pollRequestsByAgent: map[string][]assignmentcontract.PollRequest{},
	}
	f.server = httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeAssignmentServer) URL() string {
	return f.server.URL
}
