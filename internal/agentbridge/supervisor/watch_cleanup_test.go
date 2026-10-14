package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSupervisorCancelsCancellationWatchAfterTaskCompletion(t *testing.T) {
	source := newWatchCleanupSource()
	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	startWatchCleanupSupervisor(t, source, reporter, fake)

	expectStartedTask(t, reporter, source.task.ID)
	expectSignal(t, source.watchStarted, "cancellation watch did not start")
	expectProviderStarted(t, running)
	running.EmitStdout([]byte("done"))
	running.EmitExit(0, nil)
	assertWatchCleanupResult(t, reporter)
	expectSignal(t, source.watchCtxClosed, "cancellation watch context was not canceled after completion")
}

func startWatchCleanupSupervisor(
	t *testing.T,
	source *watchCleanupSource,
	reporter *reporterProbe,
	fake *process.Fake,
) *Actor {
	t.Helper()
	actor, err := New(Config{
		DaemonID:           "daemon-watch-cleanup",
		RiidoDaemonVersion: testRiidoDaemonVersion,
		Runtime:            startRuntime(t, fake),
		Source:             source,
		Reporter:           reporter,
		PollEvery:          10 * time.Millisecond,
		HeartbeatEvery:     time.Hour,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})
	return actor
}

func assertWatchCleanupResult(t *testing.T, reporter *reporterProbe) {
	t.Helper()
	res := expectTaskResult(t, reporter, "task result was not reported")
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result status = %s, want %s; result=%+v", res.Status, agentbridge.ResultCompleted, res)
	}
}
