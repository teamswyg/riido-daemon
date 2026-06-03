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

// stubAdapter is a minimal Adapter for the bridge tests.
type stubAdapter struct {
	name         string
	detected     agentbridge.DetectResult
	startCommand agentbridge.StartCommand
	seenStart    agentbridge.StartRequest
}

func (a *stubAdapter) Name() string { return a.name }
func (a *stubAdapter) Detect(_ context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return a.detected, nil
}
func (a *stubAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	a.seenStart = req
	if a.startCommand.Executable != "" {
		return a.startCommand, nil
	}
	cmd := a.startCommand
	cmd.Executable = req.Executable
	if cmd.Executable == "" {
		cmd.Executable = a.name
	}
	return cmd, nil
}
func (a *stubAdapter) NewParser() agentbridge.Parser { return &stubParser{} }
func (a *stubAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "chunk" {
		return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: string(raw.Bytes)}}}, nil, nil
	}
	return nil, nil, nil
}
func (a *stubAdapter) BlockedArgs() []string { return nil }

type stubParser struct{}

func (p *stubParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "chunk", Bytes: chunk}}, nil
}
func (p *stubParser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) { return nil, nil }
func (p *stubParser) Close() ([]agentbridge.RawEvent, error)                  { return nil, nil }

type protocolAdapter struct {
	stubAdapter
	driver agentbridge.ProtocolDriver
}

func (a *protocolAdapter) NewProtocolDriver(_ agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return a.driver, nil
}

type driverSpy struct {
	started chan struct{}
}

func (d *driverSpy) OnStart(context.Context, agentbridge.ProtocolIO) error {
	close(d.started)
	return nil
}

func (d *driverSpy) OnRaw(_ context.Context, raw agentbridge.RawEvent, _ agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "chunk" {
		return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: string(raw.Bytes)}}}, nil, nil
	}
	return nil, nil, nil
}

func (d *driverSpy) OnProcessExit(context.Context, agentbridge.ProcessExitStatus, agentbridge.ProtocolIO) ([]agentbridge.Event, error) {
	return nil, nil
}

func (d *driverSpy) OnClose(context.Context, agentbridge.ProtocolIO) error { return nil }

func TestNewRequiresAdapter(t *testing.T) {
	_, err := New(Config{})
	if err == nil {
		t.Fatal("expected error without adapters")
	}
}

func TestDetectReturnsCapabilities(t *testing.T) {
	a := &stubAdapter{name: "claude", detected: agentbridge.DetectResult{Available: true, Version: "1.0"}}
	b := &stubAdapter{name: "codex", detected: agentbridge.DetectResult{Available: false, Reason: "not in path"}}
	c, err := New(Config{Adapters: []agentbridge.Adapter{a, b}})
	if err != nil {
		t.Fatal(err)
	}
	caps, err := c.Detect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(caps) != 2 {
		t.Fatalf("want 2 caps, got %d", len(caps))
	}
	if caps[0].Provider != "claude" || !caps[0].Result.Available {
		t.Fatalf("claude detect: %+v", caps[0])
	}
	if caps[1].Provider != "codex" || caps[1].Result.Available {
		t.Fatalf("codex detect: %+v", caps[1])
	}
}

func TestRunUnknownProvider(t *testing.T) {
	a := &stubAdapter{name: "claude"}
	c, _ := New(Config{Adapters: []agentbridge.Adapter{a}})
	_, err := c.Run(context.Background(), TaskRequest{Provider: "ghost"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestRunReachesCompletion(t *testing.T) {
	a := &stubAdapter{name: "claude", detected: agentbridge.DetectResult{Available: true}}
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running // captured here so the test owns the reference

	c, _ := New(Config{
		Adapters: []agentbridge.Adapter{a},
		Process:  fake,
	})

	sess, err := c.Run(context.Background(), TaskRequest{
		ID: "t-1", Provider: "claude", Prompt: "hello",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	go func() {
		running.EmitStdout([]byte("hello"))
		running.EmitExit(0, nil)
	}()

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCompleted || res.Output != "hello" {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for result")
	}
}

func TestRunPassesDetectedExecutableToBuildStartAndSpawn(t *testing.T) {
	selected := "/opt/riido/bin/openclaw-supported"
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
		ID: "t-openclaw", Provider: "openclaw", Prompt: "hello",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if a.seenStart.Executable != selected {
		t.Fatalf("BuildStart executable = %q, want %q", a.seenStart.Executable, selected)
	}
	if got := running.Command().Executable; got != selected {
		t.Fatalf("spawn executable = %q, want %q", got, selected)
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
