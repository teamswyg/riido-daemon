package supervisor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type blockedTextReporter struct {
	*reporterProbe
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (r *blockedTextReporter) ReportEvent(ctx context.Context, task string, ev agentbridge.Event) error {
	if ev.Kind == agentbridge.EventTextDelta || ev.Kind == agentbridge.EventProgress && ev.RecoveryContext == nil {
		r.once.Do(func() { close(r.entered) })
		select {
		case <-r.release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return r.reporterProbe.ReportEvent(ctx, task, ev)
}

// Exercise the real session -> supervisor -> reporter chain with more output
// than both event buffers can hold. The provider finishes while HTTP is stuck.
func TestCompletionBypassesBlockedProgressReport(t *testing.T) {
	reporter, child, _ := startBlockedProgressRun(t, 5*time.Second)
	sent := time.Now()
	for range 700 {
		child.EmitStdout([]byte("event"))
	}
	child.EmitStdout([]byte("done"))
	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCompleted || res.Output != "done" {
			t.Fatalf("completion corrupted: %+v", res)
		}
		t.Logf("provider completion to terminal report <= %s", time.Since(sent))
	case <-time.After(time.Second):
		t.Fatal("completion blocked behind progress; expected completed result within 1s without releasing reporter")
	}
}

func startBlockedProgressRun(t *testing.T, hardTimeout time.Duration) (*blockedTextReporter, *process.FakeRunning, *controlplane.MemorySource) {
	t.Helper()
	source := controlplane.NewMemorySource()
	req := taskResultRequest()
	req.Timeout = hardTimeout
	source.Enqueue(req)
	reporter := &blockedTextReporter{reporterProbe: newReporterProbe(), entered: make(chan struct{}), release: make(chan struct{})}
	child := process.NewFakeRunning()
	rt := startRuntime(t, &process.Fake{NextRunning: child})
	a, err := New(Config{DaemonID: "daemon-1", RiidoDaemonVersion: testRiidoDaemonVersion, Runtime: rt, Source: source, Reporter: reporter, Workdir: workdir.NewFSAdapter(t.TempDir()), PollEvery: time.Millisecond, HeartbeatEvery: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		close(reporter.release)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = a.Stop(ctx)
	})
	expectTaskResultStarted(t, reporter.reporterProbe, "t-1")
	expectTaskResultRunningEvent(t, reporter.reporterProbe)
	child.EmitStdout([]byte("event"))
	select {
	case <-reporter.entered:
	case <-time.After(time.Second):
		t.Fatal("report not started")
	}
	return reporter, child, source
}

func TestBlockedProgressKeepsHardTimeoutAndCancellation(t *testing.T) {
	for _, mode := range []string{"timeout", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			reporter, _, source := startBlockedProgressRun(t, 750*time.Millisecond)
			want := agentbridge.ResultTimeout
			if mode == "cancel" {
				want = agentbridge.ResultCancelled
				source.Cancel("t-1", context.Canceled)
			}
			select {
			case res := <-reporter.results:
				if res.Status != want {
					t.Fatalf("status=%s want=%s", res.Status, want)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("actual termination stuck behind progress")
			}
		})
	}
}
