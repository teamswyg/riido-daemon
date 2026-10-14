package saasplane

import (
	"encoding/json"
	"net/http"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

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
