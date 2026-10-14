package runtimeactor

import (
	"context"
	"strconv"
	"testing"
	"time"
)

func TestRuntimeActorStopCascadesToProcesses(t *testing.T) {
	a, p := startAvailableFakeActor(t, Config{MaxConcurrent: 3})
	for i := range 3 {
		submitFakeTask(t, a, "t-"+strconv.Itoa(i))
		_ = waitForRunning(t, p, i, time.Second)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = a.Stop(ctx)
	}()

	for i := range 3 {
		r := p.at(i)
		if r == nil {
			t.Fatalf("running #%d missing", i)
		}
		expectFakeProcessKill(t, r, "process #"+strconv.Itoa(i)+" kill never received")
	}
}
