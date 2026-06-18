package saasplane

import (
	"context"
	"testing"
)

func TestPlaneSendsDeviceCredentialHeaders(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	fake.enqueue(queuedHTTPAssignment("asn-1", "hello"))
	plane, err := New(Config{
		BaseURL:      fake.URL(),
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
		Agents:       []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask with device credential: %v", err)
	}
	if req == nil || req.ID != "asn-1" {
		t.Fatalf("request = %+v", req)
	}
}
