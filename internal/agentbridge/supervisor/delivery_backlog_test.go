package supervisor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type backlogReporter struct {
	events []agentbridge.Event
	done   chan struct{}
}

func (*backlogReporter) StartTask(context.Context, string) error { return nil }
func (r *backlogReporter) ReportEvent(ctx context.Context, _ string, ev agentbridge.Event) error {
	select {
	case <-time.After(5 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}
	r.events = append(r.events, ev)
	return nil
}

func (r *backlogReporter) CompleteTask(context.Context, string, agentbridge.Result) error {
	close(r.done)
	return nil
}

func TestDeliveryBacklogDoesNotPayOneRoundTripPerToken(t *testing.T) {
	reporter := &backlogReporter{done: make(chan struct{})}
	a := &Actor{cfg: Config{Reporter: reporter, PollEvery: time.Hour, HeartbeatEvery: time.Hour}, mailbox: make(chan envelope, 128), stoppedCh: make(chan struct{}), stopErrCh: make(chan error, 1)}
	for range 20 {
		a.mailbox <- envelope{taskEvent: &taskEventMsg{taskID: "answer", event: agentbridge.Event{Kind: agentbridge.EventLog, Text: "startup notification"}}}
	}
	for range 44 {
		a.mailbox <- envelope{taskEvent: &taskEventMsg{taskID: "answer", event: agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "가"}}}
	}
	a.mailbox <- envelope{taskResult: &taskResultMsg{taskID: "answer", result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: strings.Repeat("가", 44)}}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	start := time.Now()
	go a.run(ctx, nil)
	select {
	case <-reporter.done:
	case <-time.After(3 * time.Second):
		t.Fatal("buffered answer did not finish")
	}
	cancel()
	<-a.stoppedCh
	var answer strings.Builder
	for _, ev := range reporter.events {
		if ev.Kind == agentbridge.EventTextDelta {
			answer.WriteString(ev.Text)
		}
	}
	t.Logf("network reports=%d delivery=%s", len(reporter.events), time.Since(start))
	if answer.String() != strings.Repeat("가", 44) {
		t.Fatal("answer lost or reordered")
	}
	if len(reporter.events) > 4 {
		t.Fatalf("buffered answer paid %d serial network round trips; want <=4", len(reporter.events))
	}
}
