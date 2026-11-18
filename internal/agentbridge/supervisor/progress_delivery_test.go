package supervisor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestProgressDeliveryGroupsWithoutHoldingFirstUpdate(t *testing.T) {
	seen := make(chan agentbridge.Event, 4)
	firstDone := make(chan struct{})
	d := newProgressDelivery(t.Context(), func(ctx context.Context, ev agentbridge.Event) error {
		seen <- ev
		if ev.Text == "first" {
			select {
			case <-firstDone:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})
	defer d.stop()
	d.enqueue(agentbridge.Event{Kind: agentbridge.EventProgress, Text: "first"})
	select {
	case <-seen:
	case <-time.After(time.Second):
		t.Fatal("first update held for batch")
	}
	for _, text := range []string{"검토 ", "중", "입니다"} {
		d.enqueue(agentbridge.Event{Kind: agentbridge.EventProgress, Text: text})
	}
	close(firstDone)
	select {
	case ev := <-seen:
		if ev.Text != "검토 중입니다" {
			t.Fatalf("fragmented progress: %+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("batch not reported")
	}
}

func TestProgressDeliveryBoundsRemoteBacklog(t *testing.T) {
	d := &progressDelivery{wake: make(chan struct{}, 1)}
	for i := range 1000 {
		kind := agentbridge.EventProgress
		if i%2 == 0 {
			kind = agentbridge.EventLog
		}
		d.enqueue(agentbridge.Event{Kind: kind, Text: strings.Repeat("x", 8000)})
	}
	if len(d.queue) > maxDeliveryGroupEvents || d.bytes > maxProgressQueueBytes || d.superseded == 0 {
		t.Fatalf("unbounded queue: %d groups / %d bytes", len(d.queue), d.bytes)
	}
	d.enqueue(agentbridge.Event{Kind: agentbridge.EventProgress, Text: strings.Repeat("x", maxDeliveryGroupBytes+1)})
	if d.bytes > maxProgressQueueBytes {
		t.Fatal("single oversized frame bypassed bound")
	}
}

func TestProgressDeliveryStopsBeforeCriticalBoundary(t *testing.T) {
	entered := make(chan struct{})
	exited := make(chan struct{})
	d := newProgressDelivery(t.Context(), func(ctx context.Context, _ agentbridge.Event) error {
		close(entered)
		<-ctx.Done()
		close(exited)
		return ctx.Err()
	})
	d.enqueue(agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "partial"})
	<-entered
	d.enqueue(agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "queued"})
	done := make(chan struct{})
	go func() { d.stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("HTTP cancellation did not unblock boundary")
	}
	select {
	case <-exited:
	default:
		t.Fatal("critical event could overtake still-running display report")
	}
	if len(d.queue) != 0 {
		t.Fatal("obsolete display could follow terminal")
	}
}
