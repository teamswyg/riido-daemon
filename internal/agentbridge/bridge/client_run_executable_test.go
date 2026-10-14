package bridge

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestRunPassesDetectedExecutableToBuildStartAndSpawn(t *testing.T) {
	selected := "/opt/riido/bin/openclaw-supported"
	launchPath := "/riido/test/bin"
	a := &stubAdapter{name: "openclaw", detected: agentbridge.DetectResult{
		Available:  true,
		Executable: selected,
	}}
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running

	c, _ := New(Config{
		Adapters: []agentbridge.Adapter{a},
		Process:  fake,
	})

	sess, err := c.Run(context.Background(), TaskRequest{
		ID:       "t-openclaw",
		Provider: "openclaw",
		Prompt:   "hello",
		Env:      map[string]string{"PATH": launchPath},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if a.seenStart.Executable != selected {
		t.Fatalf("BuildStart executable = %q, want %q", a.seenStart.Executable, selected)
	}
	if got := a.seenStart.Env["PATH"]; got != launchPath {
		t.Fatalf("BuildStart PATH = %q, want %q", got, launchPath)
	}
	if got := running.Command().Executable; got != selected {
		t.Fatalf("spawn executable = %q, want %q", got, selected)
	}
	if got, ok := envListValue(running.Command().Env, "PATH"); !ok || got != launchPath {
		t.Fatalf("spawn PATH = %q ok=%v, want %q; env=%v", got, ok, launchPath, running.Command().Env)
	}

	go func() {
		running.EmitStdout([]byte("ok"))
		running.EmitExit(0, nil)
	}()
	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for result")
	}
}
