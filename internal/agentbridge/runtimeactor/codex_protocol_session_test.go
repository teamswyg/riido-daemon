package runtimeactor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// TestRuntimeActorRunsCodexWithProtocolDriver is the M-9 integration
// success criterion: a TaskRequest goes through RuntimeActor.Submit →
// SessionActor → ProtocolDriver, walks the full handshake against a
// fake process, and reaches ResultCompleted WITHOUT any inline test
// driver writing to stdin.
func TestRuntimeActorRunsCodexWithProtocolDriver(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters:      []agentbridge.Adapter{codexLikeAdapter{}},
		MaxConcurrent: 1,
	})

	h, err := a.Submit(context.Background(), bridge.TaskRequest{
		ID: "t-codex", Provider: "codex-like", Prompt: "do the thing",
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	running := waitForRunning(t, p, 0, time.Second)

	// Drain events so the session loop doesn't block.
	go func() {
		for range h.Events() {
		}
	}()

	// "Codex server" side. It must observe stdin writes and emit the
	// matching JSON-RPC responses on stdout.
	driveCodexServer(t, running)

	// The session must reach ResultCompleted purely through the driver
	// path — no inline test-driver-side stdin writes here.
	select {
	case res := <-h.Result():
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("status: %s", res.Status)
		}
		if res.Output != "all done" {
			t.Fatalf("output: %q", res.Output)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("session never reached ResultCompleted via ProtocolDriver")
	}

	// Slot must release.
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		s, _ := a.Status(context.Background())
		if s.RunningSessions == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("RunningSessions never returned to 0")
}
