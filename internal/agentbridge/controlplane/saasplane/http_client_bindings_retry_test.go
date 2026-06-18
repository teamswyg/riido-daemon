package saasplane

import (
	"context"
	"net/http"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

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
	fake.enqueue(queuedHTTPAssignment("asn-1", "dynamic binding task"))
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
