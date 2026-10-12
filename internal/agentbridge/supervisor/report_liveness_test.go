package supervisor

import (
	"context"
	"errors"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/process"
	"sync"
	"testing"
	"time"
)

type heldReporter struct {
	once    sync.Once
	entered chan struct{}
	release chan struct{}
}

func (r *heldReporter) StartTask(context.Context, string) error { return nil }
func (r *heldReporter) CompleteTask(context.Context, string, agentbridge.Result) error {
	return nil
}
func (r *heldReporter) ReportEvent(ctx context.Context, _ string, _ agentbridge.Event) error {
	r.once.Do(func() { close(r.entered) })
	select {
	case <-r.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func TestSlowReportingDoesNotStarveHeartbeat(t *testing.T) {
	source := &holdClaimSource{heartbeats: make(chan struct{}, 16), release: make(chan struct{})}
	defer close(source.release)
	reporter := &heldReporter{entered: make(chan struct{}), release: make(chan struct{})}
	rt := startRuntime(t, process.NewFake())
	a, e := New(Config{DaemonID: "audit", Runtime: rt, Source: source, Reporter: reporter, HeartbeatEvery: 20 * time.Millisecond})
	if e != nil {
		t.Fatal(e)
	}
	if e = a.Start(context.Background()); e != nil {
		t.Fatal(e)
	}
	defer func() {
		close(reporter.release)
		ctx, c := context.WithTimeout(context.Background(), time.Second)
		defer c()
		_ = a.Stop(ctx)
	}()
	a.mailbox <- envelope{taskEvent: &taskEventMsg{taskID: "t", event: agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "a"}}}
	select {
	case <-reporter.entered:
	case <-time.After(time.Second):
		t.Fatal("report did not start")
	}
	for len(source.heartbeats) > 0 {
		<-source.heartbeats
	}
	select {
	case <-source.heartbeats:
	case <-time.After(150 * time.Millisecond):
		t.Fatal("20ms heartbeat stopped for 150ms while ReportEvent was held")
	}
}

func TestHeldReportingDoesNotBlockProviderCancellation(t *testing.T) {
	source := &cancelSource{req: bridge.TaskRequest{ID: "cancel-held", Provider: "fake", Prompt: "work"}, cancel: make(chan error, 1)}
	reporter := &heldReporter{entered: make(chan struct{}), release: make(chan struct{})}
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)
	actor, err := New(Config{DaemonID: "held", Runtime: rt, Source: source, Reporter: reporter, HeartbeatEvery: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if err = actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		close(reporter.release)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	}()
	select {
	case <-reporter.entered:
	case <-time.After(time.Second):
		t.Fatal("report not held")
	}
	source.cancel <- errors.New("user cancelled")
	select {
	case <-running.KillRecv():
	case <-time.After(time.Second):
		t.Fatal("provider cancellation waited for held report")
	}
}
