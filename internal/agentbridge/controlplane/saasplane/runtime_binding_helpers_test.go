package saasplane

import (
	"context"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func newRuntimeBindingPlane(t *testing.T, fake *fakeAssignmentServer, configure func(*Config)) *Plane {
	t.Helper()
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	cfg := Config{
		BaseURL:      fake.URL(),
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
	}
	if configure != nil {
		configure(&cfg)
	}
	plane, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(plane.Close)
	return plane
}

func codexRuntimeBinding(agentID string) assignmentcontract.AgentRuntimeBinding {
	return assignmentcontract.AgentRuntimeBinding{
		AgentID:         agentID,
		DaemonID:        "daemon-1",
		DeviceID:        "device-1",
		RuntimeID:       "daemon-1:codex",
		RuntimeProvider: "codex",
	}
}

func cursorRuntimeBinding(agentID string) assignmentcontract.AgentRuntimeBinding {
	binding := codexRuntimeBinding(agentID)
	binding.RuntimeID = "daemon-1:cursor"
	binding.RuntimeProvider = "cursor"
	return binding
}

func cursorRuntimeRegistration() controlplane.RuntimeRegistration {
	return controlplane.RuntimeRegistration{
		DaemonID:   "daemon-1",
		RuntimeID:  "daemon-1:cursor",
		Provider:   "cursor",
		DeviceName: "주윤의 MacBook",
	}
}

func runtimeBindingStartedAt() time.Time {
	return time.Now().Add(-5 * time.Minute).UTC()
}

func registerRuntimeForBinding(t *testing.T, plane *Plane, rt controlplane.RuntimeRegistration) {
	t.Helper()
	if err := plane.RegisterRuntime(context.Background(), rt); err != nil {
		t.Fatalf("RegisterRuntime %s: %v", rt.Provider, err)
	}
}
