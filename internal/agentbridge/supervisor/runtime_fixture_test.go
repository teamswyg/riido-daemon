package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func startRuntime(t *testing.T, fake *process.Fake) *runtimeactor.Actor {
	t.Helper()
	return startNamedRuntime(t, fake, "rt-local", "fake")
}

func startNamedRuntime(t *testing.T, fake *process.Fake, runtimeID, provider string) *runtimeactor.Actor {
	t.Helper()
	rt, err := runtimeactor.New(runtimeactor.Config{
		RuntimeID:     runtimeID,
		Owner:         "owner-a",
		DeviceName:    "device-a",
		Adapters:      []agentbridge.Adapter{&stubAdapter{name: provider}},
		Process:       fake,
		MaxConcurrent: 1,
		MailboxSize:   8,
	})
	if err != nil {
		t.Fatalf("runtime New: %v", err)
	}
	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("runtime Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = rt.Stop(ctx)
	})
	return rt
}
