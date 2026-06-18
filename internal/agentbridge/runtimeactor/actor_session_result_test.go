package runtimeactor

import (
	"testing"
	"time"
)

func TestRuntimeActorStartsSessionAndReportsResult(t *testing.T) {
	a, p := startAvailableFakeActor(t, Config{MaxConcurrent: 1})
	h := submitFakeTask(t, a, "t-1")
	r := waitForRunning(t, p, 0, time.Second)

	go func() {
		r.EmitStdout([]byte("hello"))
		r.EmitExit(0, nil)
	}()

	expectTaskOutput(t, h.Result(), "hello", "no result")
	expectRunningSessions(t, a, 0, "RunningSessions never returned to 0")
}
