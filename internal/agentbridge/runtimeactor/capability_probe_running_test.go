package runtimeactor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestUnknownAuthProbePreservesRunningSessionAndRecovers(t *testing.T) {
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	verified := claudeCapabilityDetectResult("/fake/claude")
	adapter := &mutableDetectAdapter{stubAdapter: stubAdapter{name: "claude", detected: verified}}
	actor, proc := startActor(t, Config{Adapters: []agentbridge.Adapter{adapter}, CapabilityRefreshEvery: time.Second, Now: func() time.Time { return now }})
	handle, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "ongoing", Provider: "claude"})
	if err != nil {
		t.Fatal(err)
	}
	running := waitForRunning(t, proc, 0, time.Second)
	original := actorStatusCapabilities(t, actor)[0]
	adapter.setDetected(agentbridge.DetectResult{Executable: verified.Executable, HealthStatus: hostintegration.ProviderHealthUnknown, DiagnosticCode: hostintegration.ProviderDiagnosticAuthProbeFailed, Reason: "provider authentication probe did not complete"})
	now = now.Add(2 * time.Second)
	unknown := actorStatusCapabilities(t, actor)[0]
	if unknown.CapabilityFingerprint != original.CapabilityFingerprint || unknown.HealthStatus != hostintegration.ProviderHealthUnknown {
		t.Fatalf("unknown probe changed identity: %+v", unknown)
	}
	select {
	case res := <-handle.Result():
		t.Fatalf("unknown probe stopped ongoing work: %+v", res)
	default:
	}
	if _, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "probe-unknown", Provider: "claude"}); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("unknown auth admitted new task: %v", err)
	}
	if proc.count() != 1 {
		t.Fatal("unknown auth spawned another provider")
	}
	adapter.setDetected(verified)
	now = now.Add(2 * time.Second)
	recovered := actorStatusCapabilities(t, actor)[0]
	if !recovered.Available || recovered.CapabilityFingerprint != original.CapabilityFingerprint {
		t.Fatalf("recovered capability=%+v", recovered)
	}
	select {
	case <-running.KillRecv():
		t.Fatal("probe recovery killed provider")
	default:
	}
	changed := verified
	changed.Version = "3.0.0"
	adapter.setDetected(changed)
	now = now.Add(2 * time.Second)
	_, _ = actor.Status(context.Background())
	expectFakeProcessKill(t, running, "confirmed version drift did not fence provider")
	if res := expectTaskResult(t, handle.Result(), "missing drift result"); res.Status != agentbridge.ResultBlocked {
		t.Fatalf("drift=%+v", res)
	}
}

func TestSubmitRefreshFencesAlreadyRunningSession(t *testing.T) {
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	adapter := &mutableDetectAdapter{stubAdapter: stubAdapter{name: "fake", detected: fakeDetect("1.0.0")}}
	actor, proc := startActor(t, Config{Adapters: []agentbridge.Adapter{adapter}, CapabilityRefreshEvery: time.Second, Now: func() time.Time { return now }})
	handle, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "old", Provider: "fake"})
	if err != nil {
		t.Fatal(err)
	}
	running := waitForRunning(t, proc, 0, time.Second)
	adapter.setDetected(fakeDetect("2.0.0"))
	now = now.Add(2 * time.Second)
	if _, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "new", Provider: "fake"}); err != nil {
		t.Fatal(err)
	}
	expectFakeProcessKill(t, running, "submit refresh lost existing session pin")
	if res := expectTaskResult(t, handle.Result(), "missing drift result"); res.Status != agentbridge.ResultBlocked {
		t.Fatalf("drift=%+v", res)
	}
}
