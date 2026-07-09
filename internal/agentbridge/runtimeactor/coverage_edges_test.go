package runtimeactor

import (
	"errors"
	"strings"
	"testing"
	"time"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestKnownProviderCapabilityProfiles(t *testing.T) {
	tests := []struct {
		provider string
		protocol providercap.ProtocolKind
		maturity providercap.ProtocolMaturity
		worktree bool
	}{
		{"claude_code", providercap.ProtocolClaudeStreamJSON, providercap.ProtocolMaturityStable, true},
		{"codex", providercap.ProtocolCodexAppServer, providercap.ProtocolMaturityExperimental, true},
		{"openclaw", providercap.ProtocolOpenClawAgentJSON, providercap.ProtocolMaturityExperimental, false},
		{"cursor", providercap.ProtocolCursorAgentStreamJSON, providercap.ProtocolMaturityExperimental, true},
	}
	for _, tt := range tests {
		profile := profileForProvider(tt.provider)
		if profile.protocolKind != tt.protocol ||
			profile.protocolMaturity != tt.maturity ||
			profile.supportsWorktree != tt.worktree {
			t.Fatalf("%s profile = %+v", tt.provider, profile)
		}
	}
}

func TestUnknownProviderCapabilityProfile(t *testing.T) {
	profile := profileForProvider("new-provider")
	if profile.protocolKind != "new-provider-unknown" ||
		profile.protocolMaturity != providercap.ProtocolMaturityUnknown ||
		profile.eventStreamFormat != providercap.EventStreamFormatUnknown {
		t.Fatalf("unknown profile = %+v", profile)
	}
}

func TestSubmitErrorWrappersPreserveCause(t *testing.T) {
	cause := errors.New("boom")
	adapter := &stubAdapter{name: "fake"}

	buildErr := submitBuildStartError(adapter, cause)
	if !errors.Is(buildErr, cause) || !strings.Contains(buildErr.Error(), "BuildStart fake") {
		t.Fatalf("build error = %v", buildErr)
	}

	sessionErr := submitSessionStartError(cause)
	if !errors.Is(sessionErr, cause) || !strings.Contains(sessionErr.Error(), "session.Start") {
		t.Fatalf("session error = %v", sessionErr)
	}
}

func TestSessionHandleDoneSignalsCompletion(t *testing.T) {
	a, p := startAvailableFakeActor(t, Config{})
	handle := submitFakeTask(t, a, "t-done")
	running := waitForRunning(t, p, 0, time.Second)

	emitCompletedOutput(running)
	select {
	case <-handle.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("session done did not close")
	}
	expectTaskStatus(t, handle.Result(), agentbridge.ResultCompleted, "no completed result")
}
