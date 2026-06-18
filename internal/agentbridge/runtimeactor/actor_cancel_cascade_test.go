package runtimeactor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorCancelCascadesToProcessAndSession(t *testing.T) {
	a, p := startAvailableFakeActor(t, Config{})
	h := submitFakeTask(t, a, "t-cascade")
	r := waitForRunning(t, p, 0, time.Second)

	if err := a.Cancel(context.Background(), "t-cascade", "test cascade"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	expectFakeProcessKill(t, r, "process kill never received")
	expectTaskStatus(t, h.Result(), agentbridge.ResultCancelled, "session never terminated")
	expectRunningSessions(t, a, 0, "slot never released after cascade-cancel")
}
