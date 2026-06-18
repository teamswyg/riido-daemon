package runtimeactor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestRuntimeActorSubmitAfterStop(t *testing.T) {
	a, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = a.Stop(ctx)

	_, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-late", Provider: "fake"})
	if !errors.Is(err, ErrActorStopped) {
		t.Fatalf("expected ErrActorStopped, got %v", err)
	}
}
