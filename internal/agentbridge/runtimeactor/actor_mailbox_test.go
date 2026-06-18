package runtimeactor

import (
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestRuntimeActorDefaultMailboxMatchesProviderRuntimeBackpressureSSOT(t *testing.T) {
	a, err := New(Config{
		RuntimeID: "rt-mailbox-default",
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
		Process: newFakeProcess(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := cap(a.mailbox); got != DefaultMailboxSize {
		t.Fatalf("mailbox size = %d, want %d", got, DefaultMailboxSize)
	}
}

func TestRuntimeActorMailboxBackpressure(t *testing.T) {
	a, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
		MailboxSize: 1,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := a.Submit(ctx, bridge.TaskRequest{ID: "tx", Provider: "fake"})
	if err == nil {
		t.Fatal("expected error on cancelled-ctx submit")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
