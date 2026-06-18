package runtimeactor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func lateAvailableDetectResult(version, name string) agentbridge.DetectResult {
	return agentbridge.DetectResult{
		Available:  true,
		Version:    version,
		Executable: "/usr/local/bin/" + name,
	}
}

func assertSingleAvailableCapability(t *testing.T, caps []Capability, version string) {
	t.Helper()
	if len(caps) != 1 {
		t.Fatalf("want one capability, got %+v", caps)
	}
	if !caps[0].Available || caps[0].Version != version {
		t.Fatalf("refreshed capability not projected to status: %+v", caps)
	}
}
