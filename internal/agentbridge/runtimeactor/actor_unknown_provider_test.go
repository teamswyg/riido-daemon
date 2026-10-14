package runtimeactor

import (
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestRuntimeActorRejectsUnknownProvider(t *testing.T) {
	a, p := startAvailableFakeActor(t, Config{})
	_, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-1", Provider: "ghost", Prompt: "x"})
	if !errors.Is(err, ErrUnknownProvider) {
		t.Fatalf("expected ErrUnknownProvider, got %v", err)
	}
	if p.count() != 0 {
		t.Fatalf("no process should have been spawned: %d", p.count())
	}
}
