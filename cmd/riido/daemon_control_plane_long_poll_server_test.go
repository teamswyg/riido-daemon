package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
)

func newDaemonLongPollServer(t *testing.T, pollSeen chan<- assignmentcontract.PollRequest) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/daemon/agent-bindings":
			writeDaemonRuntimeTestJSON(t, w, saasplane.AgentRuntimeBindingListResponse{
				SchemaVersion: assignmentcontract.SchemaVersion,
				Bindings: []assignmentcontract.AgentRuntimeBinding{{
					AgentID:         "agent-long",
					DaemonID:        "device-1",
					DeviceID:        "device-1",
					RuntimeID:       "device-1:codex",
					RuntimeProvider: "codex",
				}},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/agents/agent-long/poll":
			var req assignmentcontract.PollRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("poll request decode: %v", err)
			}
			pollSeen <- req
			writeDaemonRuntimeTestJSON(t, w, assignmentcontract.PollResponse{
				SchemaVersion: assignmentcontract.SchemaVersion,
				Action:        assignmentcontract.PollNone,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server
}
