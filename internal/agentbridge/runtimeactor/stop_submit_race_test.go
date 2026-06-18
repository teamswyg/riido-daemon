package runtimeactor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestStopWhileSubmitInFlight(t *testing.T) {
	for trial := range 20 {
		actor := startStoppableActor(t)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		submitDone := submitFakeTaskAsync(actor, ctx)
		stopDone := stopActorAsync(actor, ctx)

		expectRaceDone(t, submitDone, cancel, trial, "Submit blocked indefinitely")
		expectRaceDone(t, stopDone, cancel, trial, "Stop blocked indefinitely")
		cancel()
	}
}

func submitFakeTaskAsync(actor *Actor, ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = actor.Submit(ctx, bridge.TaskRequest{ID: "t-race", Provider: "fake"})
	}()
	return done
}

func stopActorAsync(actor *Actor, ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = actor.Stop(ctx)
	}()
	return done
}

func expectRaceDone(t *testing.T, done <-chan struct{}, cancel context.CancelFunc, trial int, msg string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		cancel()
		t.Fatalf("trial %d: %s", trial, msg)
	}
}
