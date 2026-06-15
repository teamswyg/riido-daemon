package runtimeactor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// --- 3. Rejects unknown provider ---

func TestRuntimeActorRejectsUnknownProvider(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})
	_, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-1", Provider: "ghost", Prompt: "x"})
	if !errors.Is(err, ErrUnknownProvider) {
		t.Fatalf("expected ErrUnknownProvider, got %v", err)
	}
	if p.count() != 0 {
		t.Fatalf("no process should have been spawned: %d", p.count())
	}
}

// --- 4. Honors MaxConcurrent slots ---

func TestRuntimeActorHonorsMaxConcurrentSlots(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
		MaxConcurrent: 1,
	})
	_, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-1", Provider: "fake"})
	if err != nil {
		t.Fatalf("first submit: %v", err)
	}
	_, err = a.Submit(context.Background(), bridge.TaskRequest{ID: "t-2", Provider: "fake"})
	if !errors.Is(err, ErrSlotExhausted) {
		t.Fatalf("expected ErrSlotExhausted, got %v", err)
	}

	// Drain
	r := waitForRunning(t, p, 0, time.Second)
	r.EmitExit(0, nil)
}

// --- 5. Starts session and reports result ---

func TestRuntimeActorStartsSessionAndReportsResult(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
		MaxConcurrent: 1,
	})
	h, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-1", Provider: "fake", Prompt: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	r := waitForRunning(t, p, 0, time.Second)

	go func() {
		r.EmitStdout([]byte("hello"))
		r.EmitExit(0, nil)
	}()

	select {
	case res := <-h.Result():
		if res.Status != agentbridge.ResultCompleted || res.Output != "hello" {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no result")
	}

	// Slot must free up.
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

func TestRuntimeActorPassesDetectedExecutableToBuildStartAndSpawn(t *testing.T) {
	selected := "/opt/riido/bin/openclaw-supported"
	launchPath := "/riido/test/bin"
	startReqCh := make(chan agentbridge.StartRequest, 1)
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{
				name: "openclaw",
				detected: agentbridge.DetectResult{
					Available:  true,
					Executable: selected,
				},
				startReqCh: startReqCh,
			},
		},
		MaxConcurrent: 1,
	})
	h, err := a.Submit(context.Background(), bridge.TaskRequest{
		ID: "t-openclaw", Provider: "openclaw", Prompt: "hi", Env: map[string]string{"PATH": launchPath},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := waitForRunning(t, p, 0, time.Second)

	select {
	case req := <-startReqCh:
		if req.Executable != selected {
			t.Fatalf("BuildStart request executable = %q, want %q", req.Executable, selected)
		}
		if got := req.Env["PATH"]; got != launchPath {
			t.Fatalf("BuildStart PATH = %q, want %q", got, launchPath)
		}
	case <-time.After(time.Second):
		t.Fatal("BuildStart request not observed")
	}
	if got := p.commandAt(0).Executable; got != selected {
		t.Fatalf("spawn executable = %q, want %q", got, selected)
	}
	if got, ok := envListValue(p.commandAt(0).Env, "PATH"); !ok || got != launchPath {
		t.Fatalf("spawn PATH = %q ok=%v, want %q; env=%v", got, ok, launchPath, p.commandAt(0).Env)
	}

	go func() {
		r.EmitStdout([]byte("ok"))
		r.EmitExit(0, nil)
	}()
	select {
	case res := <-h.Result():
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no result")
	}
}

// --- 6. Cancellation cascade ---

func TestRuntimeActorCancellationCascade(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})
	h, _ := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-1", Provider: "fake"})
	r := waitForRunning(t, p, 0, time.Second)

	if err := a.Cancel(context.Background(), "t-1", "user requested"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	// Process must receive a kill signal.
	select {
	case <-r.KillRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("process not killed")
	}
	// Session result must be cancelled.
	select {
	case res := <-h.Result():
		if res.Status != agentbridge.ResultCancelled {
			t.Fatalf("status: %s", res.Status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no result")
	}
}
