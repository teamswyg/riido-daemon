package runtimeactor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestRuntimeActorHonorsMaxConcurrentSlots(t *testing.T) {
	a, p := startAvailableFakeActor(t, Config{MaxConcurrent: 1})
	submitFakeTask(t, a, "t-1")
	_, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-2", Provider: "fake"})
	if !errors.Is(err, ErrSlotExhausted) {
		t.Fatalf("expected ErrSlotExhausted, got %v", err)
	}

	r := waitForRunning(t, p, 0, time.Second)
	r.EmitExit(0, nil)
}
