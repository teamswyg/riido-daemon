package runtimeactor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func startBlockingKillActor(t *testing.T, proc *blockingKillProcess) *Actor {
	t.Helper()
	actor, err := New(Config{
		RuntimeID: "rt-stop-escalate",
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
		Process:       proc,
		MaxConcurrent: 1,
		MailboxSize:   8,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := lifecycle.DetachedShutdown(lifecycle.ShutdownForced, time.Second)
		defer cancel()
		_ = actor.StopLifecycle(ctx)
	})
	return actor
}
