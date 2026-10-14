package runtimeactor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestStopLifecycleForcedEscalatesGracefulDrain(t *testing.T) {
	proc := newBlockingKillProcess()
	t.Cleanup(proc.unblock)
	actor := startBlockingKillActor(t, proc)
	submitBlockingKillTask(t, actor)

	graceCtx, graceCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer graceCancel()
	if err := actor.Stop(graceCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("graceful Stop error = %v, want context deadline exceeded", err)
	}
	expectKillRequest(t, proc.running, lifecycle.ShutdownGraceful, "graceful Stop")

	forcedCtx, forcedCancel := lifecycle.DetachedShutdown(lifecycle.ShutdownForced, time.Second)
	defer forcedCancel()
	if err := actor.StopLifecycle(forcedCtx); err != nil {
		t.Fatalf("forced StopLifecycle: %v", err)
	}
	expectKillLevel(t, proc.running, lifecycle.ShutdownForced, "forced StopLifecycle")
}

func submitBlockingKillTask(t *testing.T, actor *Actor) {
	t.Helper()
	submitCtx, submitCancel := context.WithTimeout(context.Background(), time.Second)
	defer submitCancel()
	if _, err := actor.Submit(submitCtx, bridge.TaskRequest{ID: "t-stuck", Provider: "fake"}); err != nil {
		t.Fatalf("Submit: %v", err)
	}
}
