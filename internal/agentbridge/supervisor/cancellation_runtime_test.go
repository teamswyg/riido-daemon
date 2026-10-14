package supervisor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestSupervisorRoutesCancellationToRuntime(t *testing.T) {
	fixture := newCancellationFixture(t, bridge.TaskRequest{
		ID: "t-cancel", Provider: "fake", Prompt: "x",
	}, "rt-local", "fake")
	startRoutingSupervisor(t, Config{
		DaemonID:       "daemon-1",
		Runtime:        fixture.runtime,
		Source:         fixture.source,
		Reporter:       fixture.reporter,
		HeartbeatEvery: time.Second,
	})
	expectStartedTask(t, fixture.reporter, "t-cancel")
	expectProviderStarted(t, fixture.running)

	fixture.source.cancel <- errors.New("human cancel")
	select {
	case <-fixture.running.KillRecv():
	case <-time.After(supervisorCancellationTestTimeout):
		t.Fatal("provider process was not killed")
	}
	res := expectTaskResult(t, fixture.reporter, "cancel result was not reported")
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("result: %+v", res)
	}
}

func TestSupervisorCancelsCancellationWatcherOnComplete(t *testing.T) {
	fixture := newCancellationFixture(t, bridge.TaskRequest{
		ID: "t-complete", Provider: "fake", Prompt: "x",
	}, "rt-local", "fake")
	fixture.source.cancel = make(chan error)
	fixture.source.watchCtxs = make(chan context.Context, 1)
	startRoutingSupervisor(t, Config{
		DaemonID:       "daemon-1",
		Runtime:        fixture.runtime,
		Source:         fixture.source,
		Reporter:       fixture.reporter,
		HeartbeatEvery: time.Second,
	})
	expectStartedTask(t, fixture.reporter, "t-complete")
	expectProviderStarted(t, fixture.running)
	watchCtx := expectCancellationWatchContext(t, fixture.source)
	completeFakeProcess(fixture.running)
	res := expectTaskResult(t, fixture.reporter, "completion was not reported")
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	expectContextDone(t, watchCtx, "cancellation watcher context was not cancelled after completion")
}
