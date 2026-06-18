package saasplane

import (
	"context"
	"net/http"
	"testing"
)

func TestPlaneRetriesTransientPollTransportError(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(queuedHTTPAssignment("asn-transport", "hello"))
	transport := &transientTransport{
		failures: 1,
		next:     fake.server.Client().Transport,
	}
	plane, err := New(Config{
		BaseURL:    fake.URL(),
		DaemonID:   "daemon-1",
		DeviceID:   "device-1",
		Agents:     []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}},
		HTTPClient: &http.Client{Transport: transport},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask should retry transient transport error: %v", err)
	}
	if req == nil || req.ID != "asn-transport" {
		t.Fatalf("request = %+v", req)
	}
	if got := fake.requestCount("/v1/agents/jykim1/poll"); got != 1 {
		t.Fatalf("server poll request count = %d, want 1 after one client-side transport failure", got)
	}
	if transport.failures != 0 {
		t.Fatalf("transport failures remaining = %d, want 0", transport.failures)
	}
}
