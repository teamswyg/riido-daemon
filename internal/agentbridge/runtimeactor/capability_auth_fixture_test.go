package runtimeactor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func startPinnedProbeFixture(t *testing.T) (*Actor, *mutableDetectAdapter, *SessionHandle, agentbridge.DetectResult) {
	t.Helper()
	exe := filepath.Join(t.TempDir(), "codex")
	if err := os.WriteFile(exe, []byte("verified provider"), 0o700); err != nil {
		t.Fatal(err)
	}
	verified := fakeDetect("codex-cli 1.0.0")
	verified.Executable = exe
	verified.HealthStatus = hostintegration.ProviderHealthHealthy
	verified.DiagnosticCode = hostintegration.ProviderDiagnosticNone
	adapter := &mutableDetectAdapter{stubAdapter: stubAdapter{name: "codex", detected: verified}}
	actor, proc := startActor(t, Config{Adapters: []agentbridge.Adapter{adapter}, CapabilityRefreshEvery: -1})
	handle, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "ongoing", Provider: "codex"})
	if err != nil {
		t.Fatal(err)
	}
	waitForRunning(t, proc, 0, time.Second)
	return actor, adapter, handle, verified
}

func refreshProbeFixture(t *testing.T, actor *Actor) Capability {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := actor.RefreshCapability(ctx, "codex"); err != nil {
		t.Fatal(err)
	}
	return actorStatusCapabilities(t, actor)[0]
}
