package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func assertTaskResultRuntimeRegistration(t *testing.T, source *controlplane.MemorySource) {
	t.Helper()
	registered := source.Registered()
	if len(registered) != 1 || registered[0].RuntimeID != "rt-local" || registered[0].DaemonID != "daemon-1" {
		t.Fatalf("registered runtime: %+v", registered)
	}
	assertTaskResultRuntimeCapabilities(t, registered[0])
}

func assertTaskResultRuntimeCapabilities(t *testing.T, registration controlplane.RegisteredRuntime) {
	t.Helper()
	if _, ok := registration.Capabilities["provider.fake.supports_mcp"]; !ok {
		t.Fatalf("registered runtime missing capability support flags: %+v", registration.Capabilities)
	}
	if _, ok := registration.Capabilities["provider.fake.supports_worktree"]; !ok {
		t.Fatalf("registered runtime missing worktree support flag: %+v", registration.Capabilities)
	}
	if registration.CapabilityAttributes["provider.fake.provider_version"] == "" {
		t.Fatalf("registered runtime missing provider version attribute: %+v", registration.CapabilityAttributes)
	}
}
