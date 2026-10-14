package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestCancellationAndDeadlinePreemptFullEventBuffer(t *testing.T) {
	for _, mode := range []string{"cancel", "hard", "idle", "context"} {
		ctx, cancel := context.WithCancel(context.Background())
		fake := process.NewFake()
		adapter := &recordingAdapter{name: "fake", parser: &recordingParser{}, translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: "first"}, {Kind: agentbridge.EventTextDelta, Text: "buffer blocked"}}, nil, nil
		}}
		cfg := Config{Adapter: adapter, Process: fake, Spawn: process.Command{Executable: "fake"}, EventBuffer: 1}
		if mode == "hard" {
			cfg.HardTimeout = 100 * time.Millisecond
		}
		if mode == "idle" {
			cfg.SemanticIdle = 100 * time.Millisecond
		}
		sess, err := Start(ctx, cfg)
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		running := sess.runningForTest()
		running.EmitStdout([]byte("burst"))
		until := time.Now().Add(time.Second)
		for len(sess.events) == 0 && time.Now().Before(until) {
			time.Sleep(time.Millisecond)
		}
		if len(sess.events) != 1 {
			cancel()
			t.Fatal("event buffer never filled")
		}
		if mode == "cancel" {
			sess.Cancel(nil)
		}
		if mode == "context" {
			cancel()
		}
		select {
		case result := <-sess.Result():
			want := agentbridge.ResultCancelled
			if mode == "hard" || mode == "idle" {
				want = agentbridge.ResultTimeout
			}
			if result.Status != want {
				cancel()
				t.Fatalf("result=%+v", result)
			}
		case <-time.After(300 * time.Millisecond):
			cancel()
			<-sess.Result()
			t.Fatalf("cancellation/deadline stalled behind a full event buffer (mode=%s)", mode)
		}
		cancel()
	}
}
