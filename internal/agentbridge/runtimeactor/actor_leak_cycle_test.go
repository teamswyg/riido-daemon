package runtimeactor

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorStopReleasesAllGoroutines(t *testing.T) {
	baseline := settledGoroutineBaseline()
	for cycle := range 3 {
		runLeakCycle(t, cycle)
	}

	final := waitGoroutineCount(baseline+3, 2*time.Second)
	if final > baseline+3 {
		t.Fatalf("goroutine leak: baseline=%d final=%d (delta=%d)", baseline, final, final-baseline)
	}
}

func runLeakCycle(t *testing.T, cycle int) {
	t.Helper()
	proc := newFakeProcess()
	a := startManualAvailableFakeActor(t, Config{
		RuntimeID:     "rt-leak",
		Process:       proc,
		MaxConcurrent: 2,
	})
	for i := range 2 {
		id := "t-" + strconv.Itoa(cycle) + "-" + strconv.Itoa(i)
		h := submitFakeTask(t, a, id)
		r := waitForRunning(t, proc, i, time.Second)
		emitCompletedOutput(r)
		expectTaskStatus(t, h.Result(), agentbridge.ResultCompleted, "task did not complete")
	}
	stopLeakActor(t, a, cycle)
}

func stopLeakActor(t *testing.T, a *Actor, cycle int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := a.Stop(ctx); err != nil {
		t.Fatalf("Stop cycle %d: %v", cycle, err)
	}
}
