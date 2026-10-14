package saasplane

import (
	"errors"
	"net/http"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func queuedHTTPAssignment(id, prompt string) assignmentcontract.Assignment {
	if prompt == "" {
		prompt = "hello"
	}
	return assignmentcontract.Assignment{
		ID:              id,
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          prompt,
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
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
