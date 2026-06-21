package runtimeactor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestRuntimeActorBlocksRunningTaskOnCapabilityDrift(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	adapter := &mutableDetectAdapter{
		stubAdapter: stubAdapter{name: "fake", detected: fakeDetect("1.0.0")},
	}
	actor, proc := startActor(t, Config{
		Adapters:               []agentbridge.Adapter{adapter},
		CapabilityRefreshEvery: time.Second,
		Now:                    func() time.Time { return now },
	})
	handle, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "t-drift", Provider: "fake"})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	running := waitForRunning(t, proc, 0, time.Second)

	adapter.setDetected(fakeDetect("2.0.0"))
	now = now.Add(2 * time.Second)
	_, _ = actor.Status(context.Background())

	expectFakeProcessKill(t, running, "runtime drift did not kill provider process")
	res := expectTaskResult(t, handle.Result(), "drift result was not reported")
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("status = %s, want %s", res.Status, agentbridge.ResultBlocked)
	}
	if !strings.Contains(res.Error, ErrRuntimePinViolated.Error()) {
		t.Fatalf("error = %q", res.Error)
	}
}

func fakeDetect(version string) agentbridge.DetectResult {
	return agentbridge.DetectResult{Available: true, Version: version, Executable: "fake"}
}

func expectTaskResult(t *testing.T, results <-chan agentbridge.Result, msg string) agentbridge.Result {
	t.Helper()
	select {
	case res := <-results:
		return res
	case <-time.After(2 * time.Second):
		t.Fatal(msg)
		return agentbridge.Result{}
	}
}
