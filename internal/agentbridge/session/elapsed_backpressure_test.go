package session

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Diagnostic probe: adapter timestamps are not an independent native model clock.
// All provider events are already available; only the consumer and clock advance.
func TestAdapterElapsedIncludesBlockedEventDelivery(t *testing.T) {
	base := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	var elapsed atomic.Int64
	r := &sessionRunner{
		ctx: context.Background(), startedAt: base,
		cfg:   Config{Now: func() time.Time { return base.Add(time.Duration(elapsed.Load())) }},
		sess:  &Session{events: make(chan agentbridge.Event, 1)},
		state: agentbridge.NewState(), telemetry: agentbridge.NewTelemetryParser(),
	}
	done := make(chan agentbridge.Result, 1)
	go func() {
		r.applyEvents([]agentbridge.Event{
			{Kind: agentbridge.EventTextDelta, Text: "first"},
			{Kind: agentbridge.EventTextDelta, Text: "second"},
			{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}},
		}, nil)
		done <- r.finalResult()
	}()
	until := time.Now().Add(time.Second)
	for len(r.sess.events) == 0 && time.Now().Before(until) {
		time.Sleep(time.Millisecond)
	}
	if len(r.sess.events) == 0 {
		t.Fatal("fixture did not fill event buffer")
	}
	select {
	case <-done:
		t.Fatal("terminal advanced despite full event buffer")
	default:
	}
	elapsed.Store(int64(5 * time.Second))
	for range 3 {
		select {
		case <-r.sess.events:
		case <-time.After(time.Second):
			t.Fatal("fixture did not drain")
		}
	}
	select {
	case result := <-done:
		if got := result.FinishedAt.Sub(result.StartedAt); got != 5*time.Second {
			t.Fatalf("adapter interval=%s; want injected consumer wait", got)
		}
		t.Log("adapter interval includes injected 5s consumer wait; no native model was called")
	case <-time.After(time.Second):
		t.Fatal("fixture result did not arrive")
	}
}
