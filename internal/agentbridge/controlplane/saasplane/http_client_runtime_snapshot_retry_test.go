package saasplane

import (
	"context"
	"net/http"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneRetriesTransientRuntimeSnapshot(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	fake.failNext("/v1/daemon/runtime-snapshot", 1, http.StatusTooManyRequests)
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

	err = plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "daemon-1:codex",
		Provider:  "codex",
	})
	if err != nil {
		t.Fatalf("RegisterRuntime should retry transient runtime snapshot: %v", err)
	}
	if got := fake.requestCount("/v1/daemon/runtime-snapshot"); got != 2 {
		t.Fatalf("runtime snapshot request count = %d, want 2", got)
	}
}
