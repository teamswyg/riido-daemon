package runtimeactor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorLeavesOpenClawWorktreeUnsupported(t *testing.T) {
	actor, _ := startActor(t, Config{
		RuntimeID: "rt-openclaw",
		Adapters:  []agentbridge.Adapter{openClawLikeAdapter()},
	})
	capability := actorStatusCapabilities(t, actor)[0]
	if capability.SupportsWorktree {
		t.Fatalf("OpenClaw must not advertise daemon-selected worktree support: %+v", capability)
	}
}

func openClawLikeAdapter() *stubAdapter {
	return &stubAdapter{name: "openclaw", detected: agentbridge.DetectResult{
		Available:         true,
		Version:           "2026.5.22",
		Executable:        "/usr/local/bin/openclaw",
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsUsage:     true,
	}}
}
