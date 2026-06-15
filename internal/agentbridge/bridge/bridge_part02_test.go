package bridge

import (
	"context"
	"errors"
	"os"
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
		ID: "t-openclaw", Provider: "openclaw", Prompt: "hello", Env: map[string]string{"PATH": launchPath},
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

func TestRunInstallsProtocolDriverProvider(t *testing.T) {
	driver := &driverSpy{started: make(chan struct{})}
	a := &protocolAdapter{
		stubAdapter: stubAdapter{name: "codex", detected: agentbridge.DetectResult{Available: true}},
		driver:      driver,
	}
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	c, _ := New(Config{
		Adapters: []agentbridge.Adapter{a},
		Process:  fake,
	})

	sess, err := c.Run(context.Background(), TaskRequest{ID: "t-driver", Provider: "codex", Prompt: "hello"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	select {
	case <-driver.started:
	case <-time.After(2 * time.Second):
		t.Fatal("protocol driver OnStart was not called")
	}

	go func() {
		running.EmitStdout([]byte("ok"))
		running.EmitExit(0, nil)
	}()

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCompleted || res.Output != "ok" {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for result")
	}
}

func TestRunPassesAdapterTempFilesToSessionCleanup(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "mcp-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	a := &stubAdapter{
		name:     "claude",
		detected: agentbridge.DetectResult{Available: true},
		startCommand: agentbridge.StartCommand{
			Executable: "claude",
			TempFiles:  []string{tempFile.Name()},
		},
	}
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running

	c, _ := New(Config{
		Adapters: []agentbridge.Adapter{a},
		Process:  fake,
	})
	sess, err := c.Run(context.Background(), TaskRequest{ID: "t-tempfile", Provider: "claude"})
	if err != nil {
		t.Fatalf("Run: %v", err)
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
	if _, err := os.Stat(tempFile.Name()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temp file should be removed by bridge-run session, stat err=%v", err)
	}
}

func TestRunCancel(t *testing.T) {
	a := &stubAdapter{name: "claude", detected: agentbridge.DetectResult{Available: true}}
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	c, _ := New(Config{Adapters: []agentbridge.Adapter{a}, Process: fake})

	sess, _ := c.Run(context.Background(), TaskRequest{ID: "t-2", Provider: "claude"})
	sess.Cancel(nil)

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCancelled {
			t.Fatalf("expected cancelled, got %s", res.Status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancel timeout")
	}
}
