package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestProviderResultDoesNotWaitForFullObserverBuffer(t *testing.T) {
	child := process.NewFakeRunning()
	adapter := &recordingAdapter{name: "fake", parser: &recordingParser{}, translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
		return []agentbridge.Event{
			{Kind: agentbridge.EventLog, Text: "slow report"},
			{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "parent review"}},
		}, nil, nil
	}}
	sess, err := Start(t.Context(), Config{Adapter: adapter, Process: &process.Fake{NextRunning: child}, Spawn: process.Command{Executable: "fake"}, EventBuffer: 1, SemanticIdle: 100 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	child.EmitStdout([]byte("result"))
	select {
	case result := <-sess.Result():
		if result.Status != agentbridge.ResultCompleted || result.Output != "parent review" {
			t.Fatalf("observer backpressure changed the provider result: %+v", result)
		}
	case <-time.After(time.Second):
		t.Fatal("provider result stalled behind the observer")
	}
	if ev := <-sess.Events(); ev.Kind != agentbridge.EventLog || ev.Text != "slow report" {
		t.Fatalf("queued log = %+v", ev)
	}
}
