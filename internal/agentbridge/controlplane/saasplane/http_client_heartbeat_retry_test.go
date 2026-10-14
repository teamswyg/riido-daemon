package saasplane

import (
	"context"
	"net/http"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneRetriesTransientHeartbeat(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.failNext("/v1/agents/jykim1/heartbeat", 1, http.StatusGatewayTimeout)
	agent := AgentBinding{AgentID: "jykim1", RuntimeProvider: "codex"}
	fake.enqueue(queuedHTTPAssignment("asn-1", "hello"))
	plane := newTestPlane(t, fake.URL(), []AgentBinding{agent})
	defer plane.Close()

	runtimeID := RuntimeIDForAgent("daemon-1", agent)
	req, err := plane.ClaimTask(context.Background(), runtimeID)
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	err = plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      runtimeID,
		RunningTaskIDs: []string{req.ID},
	})
	if err != nil {
		t.Fatalf("Heartbeat should retry transient heartbeat: %v", err)
	}
	if got := fake.requestCount("/v1/agents/jykim1/heartbeat"); got != 2 {
		t.Fatalf("heartbeat request count = %d, want 2", got)
	}
}
