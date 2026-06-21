package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func startRuntimeWithAdapter(
	t *testing.T,
	fake *process.Fake,
	runtimeID string,
	adapter agentbridge.Adapter,
) *runtimeactor.Actor {
	t.Helper()
	rt, err := runtimeactor.New(runtimeactor.Config{
		RuntimeID:              runtimeID,
		Owner:                  "owner-a",
		DeviceName:             "device-a",
		Adapters:               []agentbridge.Adapter{adapter},
		Process:                fake,
		MaxConcurrent:          1,
		MailboxSize:            8,
		CapabilityRefreshEvery: time.Nanosecond,
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
