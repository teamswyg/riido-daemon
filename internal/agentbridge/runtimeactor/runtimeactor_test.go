package runtimeactor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// --- Test scaffolding ---

type stubAdapter struct {
	name       string
	detected   agentbridge.DetectResult
	startReqCh chan agentbridge.StartRequest
}

func (a *stubAdapter) Name() string { return a.name }
func (a *stubAdapter) Detect(_ context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return a.detected, nil
}

func (a *stubAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	if a.startReqCh != nil {
		select {
		case a.startReqCh <- req:
		default:
		}
	}
	exe := req.Executable
	if exe == "" {
		exe = a.name
	}
	return agentbridge.StartCommand{Executable: exe}, nil
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

// fakeProcess is a channel-owned multi-spawn fake. Single goroutine
// (the producer loop) owns the slice; Start, at, count send messages.
// No mutex — same discipline as the runtime actor mailbox boundary.
type fakeProcess struct {
	startCh chan startReq
	atCh    chan atReq
	cmdCh   chan cmdReq
	cntCh   chan chan int
}

type startReq struct {
	cmd   process.Command
	reply chan *process.FakeRunning
}

type atReq struct {
	idx   int
	reply chan *process.FakeRunning
}

type cmdReq struct {
	idx   int
	reply chan process.Command
}

func newFakeProcess() *fakeProcess {
	f := &fakeProcess{
		startCh: make(chan startReq, 8),
		atCh:    make(chan atReq, 8),
		cmdCh:   make(chan cmdReq, 8),
		cntCh:   make(chan chan int, 4),
	}
	go f.run()
	return f
}

func (f *fakeProcess) run() {
	var produced []*process.FakeRunning
	var commands []process.Command
	for {
		select {
		case msg := <-f.startCh:
			r := process.NewFakeRunning()
			produced = append(produced, r)
			commands = append(commands, msg.cmd)
			msg.reply <- r
		case msg := <-f.atCh:
			if msg.idx >= len(produced) {
				msg.reply <- nil
			} else {
				msg.reply <- produced[msg.idx]
			}
		case msg := <-f.cmdCh:
			if msg.idx >= len(commands) {
				msg.reply <- process.Command{}
			} else {
				msg.reply <- commands[msg.idx]
			}
		case reply := <-f.cntCh:
			reply <- len(produced)
		}
	}
}

func (f *fakeProcess) Start(_ context.Context, cmd process.Command) (process.RunningProcess, error) {
	reply := make(chan *process.FakeRunning, 1)
	f.startCh <- startReq{cmd: cmd, reply: reply}
	return <-reply, nil
}

func (f *fakeProcess) at(i int) *process.FakeRunning {
	reply := make(chan *process.FakeRunning, 1)
	f.atCh <- atReq{idx: i, reply: reply}
	return <-reply
}

func (f *fakeProcess) count() int {
	reply := make(chan int, 1)
	f.cntCh <- reply
	return <-reply
}

func (f *fakeProcess) commandAt(i int) process.Command {
	reply := make(chan process.Command, 1)
	f.cmdCh <- cmdReq{idx: i, reply: reply}
	return <-reply
}

func waitForRunning(t *testing.T, p *fakeProcess, i int, d time.Duration) *process.FakeRunning {
	t.Helper()
	end := time.Now().Add(d)
	for time.Now().Before(end) {
		if r := p.at(i); r != nil {
			return r
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("fake process #%d not created within %v", i, d)
	return nil
}

func startActor(t *testing.T, cfg Config) (*Actor, *fakeProcess) {
	t.Helper()
	if cfg.Process == nil {
		cfg.Process = newFakeProcess()
	}
	if cfg.MaxConcurrent == 0 {
		cfg.MaxConcurrent = 2
	}
	if cfg.MailboxSize == 0 {
		cfg.MailboxSize = 8
	}
	if cfg.RuntimeID == "" {
		cfg.RuntimeID = "rt-test"
	}
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := a.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.Stop(ctx)
	})
	return a, cfg.Process.(*fakeProcess)
}

// --- 1. Detects capabilities on start ---

func TestRuntimeActorDetectsCapabilitiesOnStart(t *testing.T) {
	avail := &stubAdapter{name: "available", detected: agentbridge.DetectResult{Available: true, Version: "1.0", Executable: "/usr/bin/available"}}
	missing := &stubAdapter{name: "missing", detected: agentbridge.DetectResult{Available: false, Reason: "not installed"}}

	a, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{avail, missing},
	})

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status.Capabilities) != 2 {
		t.Fatalf("want 2 capabilities, got %d: %+v", len(status.Capabilities), status.Capabilities)
	}
	names := map[string]Capability{}
	for _, c := range status.Capabilities {
		names[c.Provider] = c
	}
	if !names["available"].Available || names["available"].Version != "1.0" {
		t.Fatalf("available capability: %+v", names["available"])
	}
	if names["missing"].Available || names["missing"].Reason != "not installed" {
		t.Fatalf("missing capability: %+v", names["missing"])
	}
}

func TestRuntimeActorReconcilesDetectResultToProviderCapability(t *testing.T) {
	fixedNow := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	claudeLike := &stubAdapter{name: "claude", detected: agentbridge.DetectResult{
		Available:         true,
		Version:           "2.1.150",
		Executable:        "/usr/local/bin/claude",
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsSystem:    true,
		SupportsMaxTurns:  true,
		SupportsMCP:       true,
		SupportsToolHooks: true,
		SupportsUsage:     true,
	}}
	a, _ := startActor(t, Config{
		RuntimeID:           "rt-cap",
		PolicyBundleVersion: "policy-bundle.test.v1",
		Adapters:            []agentbridge.Adapter{claudeLike},
		Now:                 func() time.Time { return fixedNow },
	})

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Capabilities) != 1 {
		t.Fatalf("capabilities: %+v", status.Capabilities)
	}
	capability := status.Capabilities[0]
	if capability.ProtocolKind != string(providercap.ProtocolClaudeStreamJSON) {
		t.Fatalf("protocol kind: %+v", capability)
	}
	if capability.AdapterID != "claude" || capability.AdapterVersion != "riido-agentbridge-adapter.v1" || capability.ProtocolVersion != "v1" {
		t.Fatalf("execution fingerprint fields missing: %+v", capability)
	}
	if capability.CompatibilityStatus != string(providercap.CompatSupported) {
		t.Fatalf("compatibility status: %+v", capability)
	}
	if capability.CapabilityFingerprint == "" {
		t.Fatalf("fingerprint missing: %+v", capability)
	}
	if !capability.SupportsStreaming || !capability.SupportsResume || !capability.SupportsSystem ||
		!capability.SupportsMaxTurns || !capability.SupportsMCP || !capability.SupportsToolHooks ||
		!capability.SupportsUsage || !capability.SupportsWorktree {
		t.Fatalf("surface flags were not preserved: %+v", capability)
	}
	if capability.SupportsFileEvents {
		t.Fatalf("file events must stay false until a provider emits structured file events: %+v", capability)
	}
}

func TestRuntimeActorLeavesOpenClawWorktreeUnsupported(t *testing.T) {
	openclawLike := &stubAdapter{name: "openclaw", detected: agentbridge.DetectResult{
		Available:         true,
		Version:           "2026.5.22",
		Executable:        "/usr/local/bin/openclaw",
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsUsage:     true,
	}}
	a, _ := startActor(t, Config{
		RuntimeID: "rt-openclaw",
		Adapters:  []agentbridge.Adapter{openclawLike},
	})

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	capability := status.Capabilities[0]
	if capability.SupportsWorktree {
		t.Fatalf("OpenClaw must not advertise daemon-selected worktree support without a native workspace surface: %+v", capability)
	}
}

func TestRuntimeActorDetectedFingerprintHashesExecutable(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "claude")
	content := []byte("provider binary v1\n")
	if err := os.WriteFile(binary, content, 0o755); err != nil {
		t.Fatal(err)
	}
	wantSum := sha256.Sum256(content)
	want := hex.EncodeToString(wantSum[:])

	claudeLike := &stubAdapter{name: "claude", detected: agentbridge.DetectResult{
		Available:         true,
		Version:           "2.1.150",
		Executable:        binary,
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsSystem:    true,
		SupportsMaxTurns:  true,
		SupportsMCP:       true,
		SupportsToolHooks: true,
		SupportsUsage:     true,
	}}
	a, _ := startActor(t, Config{
		RuntimeID:           "rt-detected-fp",
		PolicyBundleVersion: "policy-bundle.test.v1",
		Adapters:            []agentbridge.Adapter{claudeLike},
	})

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	capability := status.Capabilities[0]
	if capability.DetectedFingerprint != want {
		t.Fatalf("detected fingerprint = %q, want %q", capability.DetectedFingerprint, want)
	}
	withFingerprint := capability.CapabilityFingerprint
	if withFingerprint == "" {
		t.Fatalf("capability fingerprint missing: %+v", capability)
	}

	noBinary := *claudeLike
	noBinary.detected.Executable = "claude"
	a, _ = startActor(t, Config{
		RuntimeID:           "rt-detected-fp-empty",
		PolicyBundleVersion: "policy-bundle.test.v1",
		Adapters:            []agentbridge.Adapter{&noBinary},
	})
	status, err = a.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Capabilities[0].DetectedFingerprint != "" {
		t.Fatalf("non-absolute executable must not be fingerprinted: %+v", status.Capabilities[0])
	}
	if status.Capabilities[0].CapabilityFingerprint == withFingerprint {
		t.Fatal("capability fingerprint must include detected fingerprint input")
	}
}

func TestRuntimeActorReconcilesUnavailableProviderAsBlocked(t *testing.T) {
	a, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "cursor", detected: agentbridge.DetectResult{
				Available: false,
				Reason:    "cursor-agent missing",
			}},
		},
	})

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	capability := status.Capabilities[0]
	if capability.Available {
		t.Fatalf("capability should be unavailable: %+v", capability)
	}
	if capability.CompatibilityStatus != string(providercap.CompatBlocked) {
		t.Fatalf("unavailable provider must be blocked: %+v", capability)
	}
	if capability.ProtocolKind != string(providercap.ProtocolCursorAgentStreamJSON) {
		t.Fatalf("cursor protocol kind missing: %+v", capability)
	}
}

func TestRuntimeActorCapabilityFingerprintIncludesPolicyBundle(t *testing.T) {
	detected := agentbridge.DetectResult{
		Available:         true,
		Version:           "2.1.150",
		Executable:        "/usr/local/bin/claude",
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsSystem:    true,
		SupportsMaxTurns:  true,
		SupportsMCP:       true,
		SupportsToolHooks: true,
		SupportsUsage:     true,
	}
	start := func(policy string) string {
		t.Helper()
		a, _ := startActor(t, Config{
			RuntimeID:           "rt-policy",
			PolicyBundleVersion: policy,
			Adapters:            []agentbridge.Adapter{&stubAdapter{name: "claude", detected: detected}},
		})
		status, err := a.Status(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		return status.Capabilities[0].CapabilityFingerprint
	}

	v1 := start("policy-bundle.test.v1")
	v2 := start("policy-bundle.test.v2")
	if v1 == "" || v2 == "" {
		t.Fatalf("fingerprint missing: v1=%q v2=%q", v1, v2)
	}
	if v1 == v2 {
		t.Fatal("capability fingerprint must change when policy bundle version changes")
	}
}

// --- 2. daemon status populates runtimes from Actor ---
// This is exercised in cmd/riido/daemon_test.go after wiring; here we
// verify the JSON shape is producible from Actor.Status alone.

func TestRuntimeActorStatusJSONShape(t *testing.T) {
	a, _ := startActor(t, Config{
		Owner:      "kim",
		DeviceName: "MacBook-Pro-SK.local",
		Agents: []AgentStatus{
			{AgentID: "riido", Name: "Riido", State: "online"},
		},
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})
	status, _ := a.Status(context.Background())

	// Required runtime status fields per provider-runtime SSOT.
	if status.RuntimeID == "" {
		t.Fatal("RuntimeID empty")
	}
	if status.Health != "ok" {
		t.Fatalf("Health: %q", status.Health)
	}
	if status.StartedAt.IsZero() {
		t.Fatal("StartedAt zero")
	}
	if status.MaxConcurrent == 0 {
		t.Fatal("MaxConcurrent zero")
	}
	if status.Owner != "kim" || status.DeviceName != "MacBook-Pro-SK.local" {
		t.Fatalf("Figma runtime fields: owner=%q device=%q", status.Owner, status.DeviceName)
	}
	if len(status.Agents) != 1 || status.Agents[0].Name != "Riido" {
		t.Fatalf("Agents: %+v", status.Agents)
	}
}

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
	h, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-openclaw", Provider: "openclaw", Prompt: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	r := waitForRunning(t, p, 0, time.Second)

	select {
	case req := <-startReqCh:
		if req.Executable != selected {
			t.Fatalf("BuildStart request executable = %q, want %q", req.Executable, selected)
		}
	case <-time.After(time.Second):
		t.Fatal("BuildStart request not observed")
	}
	if got := p.commandAt(0).Executable; got != selected {
		t.Fatalf("spawn executable = %q, want %q", got, selected)
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

func TestRuntimeActorTaskStatusIncluded(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})
	_, _ = a.Submit(context.Background(), bridge.TaskRequest{ID: "t-7", Provider: "fake"})
	_ = waitForRunning(t, p, 0, time.Second)

	s, _ := a.Status(context.Background())
	if len(s.RunningTasks) != 1 {
		t.Fatalf("RunningTasks: %+v", s.RunningTasks)
	}
	if s.RunningTasks[0].TaskID != "t-7" || s.RunningTasks[0].Provider != "fake" {
		t.Fatalf("RunningTasks entry: %+v", s.RunningTasks[0])
	}
	_ = strconv.Itoa // satisfy import if unused
}
