package runtimeactor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorCancellationCascade(t *testing.T) {
	a, p := startAvailableFakeActor(t, Config{})
	h := submitFakeTask(t, a, "t-1")
	r := waitForRunning(t, p, 0, time.Second)

	if err := a.Cancel(context.Background(), "t-1", "user requested"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	expectFakeProcessKill(t, r, "process not killed")
	expectTaskStatus(t, h.Result(), agentbridge.ResultCancelled, "no result")
}
