package runtimeactor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestRuntimeActorShutdownCancelsRunningSessions(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
		MaxConcurrent: 2,
	})
	h1, _ := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-1", Provider: "fake"})
	h2, _ := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-2", Provider: "fake"})
	_ = waitForRunning(t, p, 0, time.Second)
	_ = waitForRunning(t, p, 1, time.Second)

	stopErr := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		stopErr <- a.Stop(ctx)
	}()

	for _, ch := range []<-chan agentbridge.Result{h1.Result(), h2.Result()} {
		select {
		case res := <-ch:
			if res.Status != agentbridge.ResultCancelled {
				t.Fatalf("status: %s", res.Status)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("session not terminated")
		}
	}
	if err := <-stopErr; err != nil {
		t.Fatalf("Stop: %v", err)
	}
}
