package runtimeactor

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestRuntimeActorCancelUnknownTask(t *testing.T) {
	a, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})
	err := a.Cancel(context.Background(), "ghost", "")
	if !errors.Is(err, ErrUnknownTask) {
		t.Fatalf("expected ErrUnknownTask, got %v", err)
	}
}

// --- 7. Shutdown cancels running sessions ---

func TestRuntimeActorShutdownCancelsRunningSessions(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
		MaxConcurrent: 2,
	})
	h1, _ := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-1", Provider: "fake"})
	h2, _ := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-2", Provider: "fake"})
	_ = waitForRunning(t, p, 0, time.Second)
	_ = waitForRunning(t, p, 1, time.Second)

	stopErr := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		stopErr <- a.Stop(ctx)
	}()

	for _, ch := range []<-chan agentbridge.Result{h1.Result(), h2.Result()} {
		select {
		case res := <-ch:
			if res.Status != agentbridge.ResultCancelled {
				t.Fatalf("status: %s", res.Status)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("session not terminated")
		}
	}
	if err := <-stopErr; err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

// --- 8. Heartbeat payload ---

func TestRuntimeActorHeartbeatPayload(t *testing.T) {
	a, p := startActor(t, Config{
		RuntimeID:  "rt-42",
		DeviceName: "device-a",
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true, Version: "1.0"}},
		},
		MaxConcurrent: 3,
	})
	_, _ = a.Submit(context.Background(), bridge.TaskRequest{ID: "t-1", Provider: "fake"})
	_ = waitForRunning(t, p, 0, time.Second)

	hb, err := a.HeartbeatPayload(context.Background())
	if err != nil {
		t.Fatalf("HeartbeatPayload: %v", err)
	}
	if hb.RuntimeID != "rt-42" {
		t.Fatalf("id: %q", hb.RuntimeID)
	}
	if hb.DeviceName != "device-a" {
		t.Fatalf("device name: %q", hb.DeviceName)
	}
	if hb.SlotLimit != 3 || hb.SlotsInUse != 1 {
		t.Fatalf("slots: %+v", hb)
	}
	if len(hb.RunningTaskIDs) != 1 || hb.RunningTaskIDs[0] != "t-1" {
		t.Fatalf("running ids: %v", hb.RunningTaskIDs)
	}
}

// --- 9. No provider-specific FSM ---

func TestRuntimeActorDoesNotCreateProviderSpecificFSM(t *testing.T) {
	for _, s := range agentbridge.AllStates() {
		lower := strings.ToLower(string(s))
		for _, p := range []string{"claude", "codex", "openclaw", "cursor"} {
			if strings.Contains(lower, p) {
				t.Fatalf("agentbridge RunState %q leaked provider name", s)
			}
		}
	}
}

// --- 10. Mailbox backpressure ---

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

	// Saturate the actor by submitting with an already-expired context;
	// the actor should reject promptly with ctx.Err.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	_, err := a.Submit(ctx, bridge.TaskRequest{ID: "tx", Provider: "fake"})
	if err == nil {
		t.Fatal("expected error on cancelled-ctx submit")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// --- Validation tests ---

func TestNewRequiresRuntimeID(t *testing.T) {
	_, err := New(Config{Adapters: []agentbridge.Adapter{&stubAdapter{name: "x"}}, Process: newFakeProcess()})
	if err == nil {
		t.Fatal("expected error without RuntimeID")
	}
}

func TestNewRequiresAtLeastOneAdapter(t *testing.T) {
	_, err := New(Config{RuntimeID: "rt-1", Process: newFakeProcess()})
	if err == nil {
		t.Fatal("expected error without adapters")
	}
}

func TestNewRequiresProcessPort(t *testing.T) {
	_, err := New(Config{RuntimeID: "rt-1", Adapters: []agentbridge.Adapter{&stubAdapter{name: "x"}}})
	if err == nil {
		t.Fatal("expected error without Process")
	}
}
