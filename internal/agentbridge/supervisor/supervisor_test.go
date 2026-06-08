package supervisor

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type stubAdapter struct {
	name string
}

func (a *stubAdapter) Name() string { return a.name }
func (a *stubAdapter) Detect(context.Context, agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true, Version: "1.0", Executable: a.name}, nil
}
func (a *stubAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	cmd := agentbridge.StartCommand{Executable: a.name}
	if version := req.Metadata[MetadataNativeConfigVersion]; version != "" {
		cmd.Env = append(cmd.Env, "TEST_NATIVE_CONFIG_VERSION="+version)
	}
	if nativeConfigHome := req.Metadata[MetadataNativeConfigHome]; nativeConfigHome != "" {
		cmd.Env = append(cmd.Env, "TEST_NATIVE_CONFIG_HOME="+nativeConfigHome)
	}
	return cmd, nil
}
func (a *stubAdapter) NewParser() agentbridge.Parser { return &stubParser{} }
func (a *stubAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "event" {
		return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)}}, nil, nil
	}
	if raw.Type == "chunk" {
		return []agentbridge.Event{
			{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)},
			{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: string(raw.Bytes)}},
		}, nil, nil
	}
	return nil, nil, nil
}
func (a *stubAdapter) BlockedArgs() []string { return nil }

type stubParser struct{}

func (p *stubParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	if string(chunk) == "event" {
		return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "event", Bytes: chunk}}, nil
	}
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "chunk", Bytes: chunk}}, nil
}
func (p *stubParser) FeedStderr([]byte) ([]agentbridge.RawEvent, error) { return nil, nil }
func (p *stubParser) Close() ([]agentbridge.RawEvent, error)            { return nil, nil }

type reporterProbe struct {
	started chan string
	events  chan agentbridge.Event
	results chan agentbridge.Result
}

func newReporterProbe() *reporterProbe {
	return &reporterProbe{
		started: make(chan string, 4),
		events:  make(chan agentbridge.Event, 8),
		results: make(chan agentbridge.Result, 4),
	}
}

func readNativeConfigManifest(t *testing.T, path string) workdir.NativeConfigManifest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read native config manifest: %v", err)
	}
	var manifest workdir.NativeConfigManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode native config manifest: %v", err)
	}
	return manifest
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func (r *reporterProbe) StartTask(_ context.Context, taskID string) error {
	r.started <- taskID
	return nil
}
func (r *reporterProbe) ReportEvent(_ context.Context, _ string, ev agentbridge.Event) error {
	r.events <- ev
	return nil
}
func (r *reporterProbe) CompleteTask(_ context.Context, _ string, res agentbridge.Result) error {
	r.results <- res
	return nil
}

func startRuntime(t *testing.T, fake *process.Fake) *runtimeactor.Actor {
	t.Helper()
	return startNamedRuntime(t, fake, "rt-local", "fake")
}

func startNamedRuntime(t *testing.T, fake *process.Fake, runtimeID string, provider string) *runtimeactor.Actor {
	t.Helper()
	rt, err := runtimeactor.New(runtimeactor.Config{
		RuntimeID:     runtimeID,
		Owner:         "owner-a",
		DeviceName:    "device-a",
		Adapters:      []agentbridge.Adapter{&stubAdapter{name: provider}},
		Process:       fake,
		MaxConcurrent: 1,
		MailboxSize:   8,
	})
	if err != nil {
		t.Fatalf("runtime New: %v", err)
	}
	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("runtime Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = rt.Stop(ctx)
	})
	return rt
}

type runtimeRoutingSource struct {
	claims       map[string][]bridge.TaskRequest
	registered   chan controlplane.RuntimeRegistration
	deregistered chan string
}

func newRuntimeRoutingSource(claims map[string][]bridge.TaskRequest) *runtimeRoutingSource {
	return &runtimeRoutingSource{
		claims:       claims,
		registered:   make(chan controlplane.RuntimeRegistration, 8),
		deregistered: make(chan string, 8),
	}
}

func (s *runtimeRoutingSource) RegisterRuntime(_ context.Context, rt controlplane.RuntimeRegistration) error {
	s.registered <- rt
	return nil
}

func (s *runtimeRoutingSource) DeregisterRuntime(_ context.Context, runtimeID string) error {
	s.deregistered <- runtimeID
	return nil
}

func (s *runtimeRoutingSource) Heartbeat(_ context.Context, _ controlplane.RuntimeHeartbeat) error {
	return nil
}

func (s *runtimeRoutingSource) ClaimTask(_ context.Context, runtimeID string) (*bridge.TaskRequest, error) {
	queue := s.claims[runtimeID]
	if len(queue) == 0 {
		return nil, nil
	}
	req := queue[0]
	s.claims[runtimeID] = queue[1:]
	return &req, nil
}

func (s *runtimeRoutingSource) WatchCancellation(_ context.Context, _ string) (<-chan error, error) {
	return make(chan error), nil
}

type idlePollSource struct {
	claims chan string
}

func (s *idlePollSource) RegisterRuntime(context.Context, controlplane.RuntimeRegistration) error {
	return nil
}
func (s *idlePollSource) DeregisterRuntime(context.Context, string) error { return nil }
func (s *idlePollSource) Heartbeat(context.Context, controlplane.RuntimeHeartbeat) error {
	return nil
}
func (s *idlePollSource) ClaimTask(_ context.Context, runtimeID string) (*bridge.TaskRequest, error) {
	select {
	case s.claims <- runtimeID:
	default:
	}
	return nil, nil
}
func (s *idlePollSource) WatchCancellation(context.Context, string) (<-chan error, error) {
	return make(chan error), nil
}

func TestSupervisorDefaultMailboxMatchesProviderRuntimeBackpressureSSOT(t *testing.T) {
	rt := startRuntime(t, process.NewFake())
	actor, err := New(Config{
		DaemonID: "daemon-mailbox-default",
		Runtime:  rt,
		Source:   newRuntimeRoutingSource(nil),
		Reporter: newReporterProbe(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := cap(actor.mailbox); got != DefaultMailboxSize {
		t.Fatalf("mailbox size = %d, want %d", got, DefaultMailboxSize)
	}
}

func TestSupervisorBacksOffPollingWhenIdle(t *testing.T) {
	source := &idlePollSource{claims: make(chan string, 16)}
	reporter := newReporterProbe()
	rt := startRuntime(t, process.NewFake())
	actor, err := New(Config{
		DaemonID:       "daemon-1",
		Runtime:        rt,
		Source:         source,
		Reporter:       reporter,
		PollEvery:      10 * time.Millisecond,
		IdlePollEvery:  120 * time.Millisecond,
		HeartbeatEvery: time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatalf("supervisor Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	select {
	case <-source.claims:
	case <-time.After(time.Second):
		t.Fatal("first poll did not happen")
	}
	select {
	case runtimeID := <-source.claims:
		t.Fatalf("idle poll happened before backoff elapsed: %s", runtimeID)
	case <-time.After(50 * time.Millisecond):
	}
	select {
	case <-source.claims:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("idle poll did not resume after backoff interval")
	}
}

func drainTextDelta(t *testing.T, mailbox <-chan envelope, d time.Duration) string {
	t.Helper()
	select {
	case msg := <-mailbox:
		if msg.taskEvent == nil || msg.taskEvent.event.Kind != agentbridge.EventTextDelta {
			t.Fatalf("want a text-delta task event, got %+v", msg)
		}
		return msg.taskEvent.event.Text
	case <-time.After(d):
		t.Fatal("timed out waiting for a coalesced text delta")
		return ""
	}
}

// The per-task forwarder coalesces token-level text deltas into fewer, larger
// chunks: by the max-interval timer when under the size threshold, by size, by a
// boundary (non-text) event, and on terminal.
func TestForwardSessionCoalescesTextDeltas(t *testing.T) {
	a := &Actor{
		cfg:       Config{TextFlushBytes: 1000, TextFlushInterval: 40 * time.Millisecond},
		mailbox:   make(chan envelope, 32),
		stoppedCh: make(chan struct{}),
	}
	events := make(chan agentbridge.Event, 16)
	results := make(chan agentbridge.Result, 1)
	go a.forwardSession("t-1", events, results)

	// Three small deltas (under the 1000B size threshold) coalesce into one,
	// flushed by the 40ms max-interval timer.
	events <- agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "Hel"}
	events <- agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "lo, "}
	events <- agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "world"}
	if got := drainTextDelta(t, a.mailbox, time.Second); got != "Hello, world" {
		t.Fatalf("coalesced text = %q, want %q", got, "Hello, world")
	}

	// A non-text event flushes buffered text first (ordering), then forwards.
	events <- agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "more"}
	events <- agentbridge.Event{Kind: agentbridge.EventProgress, Text: "status"}
	if got := drainTextDelta(t, a.mailbox, time.Second); got != "more" {
		t.Fatalf("flush-before-nontext = %q, want %q", got, "more")
	}
	if msg := <-a.mailbox; msg.taskEvent == nil || msg.taskEvent.event.Kind != agentbridge.EventProgress {
		t.Fatalf("want progress event after flush, got %+v", msg)
	}

	// Terminal: remaining buffered text is flushed before the result.
	events <- agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "tail"}
	close(events)
	results <- agentbridge.Result{Status: agentbridge.ResultCompleted}
	if got := drainTextDelta(t, a.mailbox, time.Second); got != "tail" {
		t.Fatalf("terminal flush = %q, want %q", got, "tail")
	}
	if final := <-a.mailbox; final.taskResult == nil || final.taskResult.result.Status != agentbridge.ResultCompleted {
		t.Fatalf("want completed result, got %+v", final)
	}
}

// A byte threshold flushes mid-stream without waiting for the timer.
func TestForwardSessionFlushesOnSizeThreshold(t *testing.T) {
	a := &Actor{
		cfg:       Config{TextFlushBytes: 5, TextFlushInterval: time.Hour},
		mailbox:   make(chan envelope, 16),
		stoppedCh: make(chan struct{}),
	}
	events := make(chan agentbridge.Event, 8)
	results := make(chan agentbridge.Result, 1)
	go a.forwardSession("t-size", events, results)

	events <- agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "abc"}
	events <- agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "defg"} // 3+4=7 >= 5
	if got := drainTextDelta(t, a.mailbox, time.Second); got != "abcdefg" {
		t.Fatalf("size-flushed text = %q, want %q", got, "abcdefg")
	}
	close(events)
	results <- agentbridge.Result{Status: agentbridge.ResultCompleted}
}

// With coalescing disabled (both knobs zero) each delta is forwarded as-is.
func TestForwardSessionPassthroughWhenDisabled(t *testing.T) {
	a := &Actor{
		cfg:       Config{},
		mailbox:   make(chan envelope, 8),
		stoppedCh: make(chan struct{}),
	}
	events := make(chan agentbridge.Event, 8)
	results := make(chan agentbridge.Result, 1)
	go a.forwardSession("t-off", events, results)

	events <- agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "a"}
	events <- agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "b"}
	if got := drainTextDelta(t, a.mailbox, time.Second); got != "a" {
		t.Fatalf("passthrough delta 1 = %q, want a", got)
	}
	if got := drainTextDelta(t, a.mailbox, time.Second); got != "b" {
		t.Fatalf("passthrough delta 2 = %q, want b", got)
	}
	close(events)
	results <- agentbridge.Result{Status: agentbridge.ResultCompleted}
}

func TestWorkdirHasWorkContent(t *testing.T) {
	dir := t.TempDir()
	if workdirHasWorkContent(dir) {
		t.Fatal("empty workdir should have no work content")
	}
	// Only daemon-injected config -> still no work content.
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".riido"), 0o755); err != nil {
		t.Fatal(err)
	}
	if workdirHasWorkContent(dir) {
		t.Fatal("config-only workdir should have no work content")
	}
	// A real source file -> has work content (e.g. a mounted repo).
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !workdirHasWorkContent(dir) {
		t.Fatal("workdir with a source file should have work content")
	}
}

func TestNoRepoWorkdirGuidance(t *testing.T) {
	g := noRepoWorkdirGuidance("/tmp/ws/runs/asn-1/workdir")
	for _, want := range []string{
		"/tmp/ws/runs/asn-1/workdir",
		"no source repository",
		"needs a codebase",
		"create a new one",
	} {
		if !strings.Contains(g, want) {
			t.Fatalf("guidance missing %q:\n%s", want, g)
		}
	}
}

// holdClaimSource blocks ClaimTask (like a long-poll holding the connection)
// until released, and signals every heartbeat.
type holdClaimSource struct {
	heartbeats chan struct{}
	release    chan struct{}
}

func (s *holdClaimSource) RegisterRuntime(context.Context, controlplane.RuntimeRegistration) error {
	return nil
}
func (s *holdClaimSource) DeregisterRuntime(context.Context, string) error { return nil }
func (s *holdClaimSource) Heartbeat(context.Context, controlplane.RuntimeHeartbeat) error {
	select {
	case s.heartbeats <- struct{}{}:
	default:
	}
	return nil
}
func (s *holdClaimSource) ClaimTask(ctx context.Context, _ string) (*bridge.TaskRequest, error) {
	select {
	case <-s.release:
	case <-ctx.Done():
	}
	return nil, nil
}
func (s *holdClaimSource) WatchCancellation(context.Context, string) (<-chan error, error) {
	return make(chan error), nil
}

// A held claim poll runs on its own goroutine, so the heartbeat loop (which
// keeps assignment leases alive) must keep firing while ClaimTask blocks.
func TestSupervisorHeartbeatNotStarvedByHeldClaim(t *testing.T) {
	source := &holdClaimSource{heartbeats: make(chan struct{}, 16), release: make(chan struct{})}
	defer close(source.release)
	rt := startRuntime(t, process.NewFake())
	actor, err := New(Config{
		DaemonID:       "daemon-hold",
		Runtime:        rt,
		Source:         source,
		Reporter:       newReporterProbe(),
		PollEvery:      10 * time.Millisecond,
		HeartbeatEvery: 30 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatalf("supervisor Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	// ClaimTask is parked; heartbeats must still fire repeatedly.
	for i := 0; i < 3; i++ {
		select {
		case <-source.heartbeats:
		case <-time.After(time.Second):
			t.Fatalf("heartbeat %d did not fire while a claim was held", i)
		}
	}
}

func TestSupervisorClaimsTaskAndReportsResult(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:       "t-1",
		Provider: "fake",
		Prompt:   "hello",
		Metadata: map[string]string{
			MetadataWorkspaceID:                   "ws-1",
			MetadataAgentName:                     "Riido",
			agentbridge.MetadataTelemetryContract: agentbridge.TelemetryPlacementPrompt,
		},
	})

	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)

	actor, err := New(Config{
		DaemonID:           "daemon-1",
		RiidoDaemonVersion: "riido-agentd v1.2.3",
		Runtime:            rt,
		Source:             source,
		Reporter:           reporter,
		Workdir:            workdir.NewFSAdapter(t.TempDir()),
		PollEvery:          10 * time.Millisecond,
		HeartbeatEvery:     time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatalf("supervisor Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	select {
	case taskID := <-reporter.started:
		if taskID != "t-1" {
			t.Fatalf("started task: %q", taskID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	select {
	case ev := <-reporter.events:
		if ev.Kind != agentbridge.EventLifecycle || ev.Phase != agentbridge.StateRunning {
			t.Fatalf("running event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("running event was not reported")
	}

	go func() {
		running.EmitStdout([]byte("done"))
		running.EmitExit(0, nil)
	}()

	select {
	case ev := <-reporter.events:
		if ev.Kind != agentbridge.EventTextDelta || ev.Text != "done" {
			t.Fatalf("event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("event was not reported")
	}

	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCompleted || res.Output != "done" {
			t.Fatalf("result: %+v", res)
		}
		if res.Workdir == "" {
			t.Fatalf("expected isolated workdir in result: %+v", res)
		}
		if running.Command().Dir != res.Workdir {
			t.Fatalf("spawn dir %q != result workdir %q", running.Command().Dir, res.Workdir)
		}
		if !hasEnvPrefix(running.Command().Env, "TEST_NATIVE_CONFIG_VERSION=") {
			t.Fatalf("native config version was not passed to adapter metadata: %+v", running.Command())
		}
		if _, err := os.Stat(filepath.Join(res.Workdir, "AGENTS.md")); err != nil {
			t.Fatalf("runtime config not injected: %v", err)
		}
		nativeConfig, err := os.ReadFile(filepath.Join(res.Workdir, "AGENTS.md"))
		if err != nil {
			t.Fatalf("read runtime config: %v", err)
		}
		if !strings.Contains(string(nativeConfig), "<riido_log>") {
			t.Fatalf("runtime config missing telemetry hard rule:\n%s", nativeConfig)
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(res.Workdir), "native-config", "AGENTS.md")); err != nil {
			t.Fatalf("native config copy not injected: %v", err)
		}
		manifestBytes, err := os.ReadFile(filepath.Join(res.Workdir, workdir.NativeConfigManifestPath))
		if err != nil {
			t.Fatalf("native config manifest not injected: %v", err)
		}
		var manifest workdir.NativeConfigManifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			t.Fatalf("decode native config manifest: %v", err)
		}
		if manifest.ProviderKind != "fake" ||
			manifest.ProtocolKind != "fake-unknown" ||
			manifest.PrimaryInstructionFile != "AGENTS.md" ||
			manifest.TelemetryContractPlacement != agentbridge.TelemetryPlacementPrompt ||
			manifest.HookMode != workdir.NativeConfigHookModeInstructionOnly {
			t.Fatalf("native config manifest = %+v", manifest)
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(res.Workdir), "native-config", filepath.FromSlash(workdir.NativeConfigManifestPath))); err != nil {
			t.Fatalf("native config manifest copy not injected: %v", err)
		}
		archive, err := os.ReadFile(filepath.Join(filepath.Dir(res.Workdir), "archive.json"))
		if err != nil {
			t.Fatalf("archive manifest not written: %v", err)
		}
		for _, want := range []string{`"schema_version": "riido-workdir-archive.v1"`, `"retention_mode": "keep-in-place"`, `"result_status": "completed"`} {
			if !strings.Contains(string(archive), want) {
				t.Fatalf("archive manifest missing %q:\n%s", want, archive)
			}
		}
		events := readRunEvents(t, filepath.Join(filepath.Dir(res.Workdir), "ir", "events.jsonl"))
		assertRunEvent(t, events, ir.EventWorkdirCreated, func(ev ir.CanonicalEvent) {
			if ev.NativeConfigVersion != "" {
				t.Fatalf("WorkdirCreated must remain pre-execute without NCV: %+v", ev)
			}
			if ev.RiidoDaemonVersion != "riido-agentd v1.2.3" {
				t.Fatalf("daemon version not stamped: %+v", ev)
			}
		})
		assertRunEvent(t, events, ir.EventNativeConfigInjected, func(ev ir.CanonicalEvent) {
			if ev.NativeConfigVersion == "" {
				t.Fatalf("NativeConfigInjected missing NCV: %+v", ev)
			}
		})
		assertRunEvent(t, events, ir.EventTextDelta, func(ev ir.CanonicalEvent) {
			if ev.NativeConfigVersion == "" {
				t.Fatalf("TextDelta missing NCV: %+v", ev)
			}
			if ev.ActorKind != ir.ActorAgent || ev.ActorID != "t-1" {
				t.Fatalf("provider event attribution mismatch: %+v", ev)
			}
			if ev.Payload["text"] != "done" {
				t.Fatalf("TextDelta payload mismatch: %+v", ev.Payload)
			}
		})
		assertRunEvent(t, events, ir.EventRunReportedDone, func(ev ir.CanonicalEvent) {
			if ev.NativeConfigVersion == "" {
				t.Fatalf("RunReportedDone missing NCV: %+v", ev)
			}
			if ev.FSMVersion != task.FSMSchemaVersion {
				t.Fatalf("RunReportedDone FSMVersion = %d, want %d", ev.FSMVersion, task.FSMSchemaVersion)
			}
			if ev.ActorKind != ir.ActorDaemon {
				t.Fatalf("RunReportedDone must be daemon-attributed: %+v", ev)
			}
		})
		assertRunEvent(t, events, ir.EventWorkdirArchived, func(ev ir.CanonicalEvent) {
			if ev.NativeConfigVersion == "" {
				t.Fatalf("WorkdirArchived missing NCV: %+v", ev)
			}
		})
	case <-time.After(2 * time.Second):
		t.Fatal("result was not reported")
	}

	registered := source.Registered()
	if len(registered) != 1 || registered[0].RuntimeID != "rt-local" || registered[0].DaemonID != "daemon-1" {
		t.Fatalf("registered runtime: %+v", registered)
	}
	if _, ok := registered[0].Capabilities["provider.fake.supports_mcp"]; !ok {
		t.Fatalf("registered runtime missing capability support flags: %+v", registered[0].Capabilities)
	}
	if _, ok := registered[0].Capabilities["provider.fake.supports_worktree"]; !ok {
		t.Fatalf("registered runtime missing worktree support flag: %+v", registered[0].Capabilities)
	}
}

func TestSupervisorDispatchesTaskToSelectedRuntimeActor(t *testing.T) {
	source := newRuntimeRoutingSource(map[string][]bridge.TaskRequest{
		"rt-codex": {{
			ID:                       "t-codex",
			Provider:                 "codex",
			Prompt:                   "hello",
			AllowExperimentalRuntime: true,
			Metadata: map[string]string{
				MetadataWorkspaceID: "ws-1",
			},
		}},
	})
	reporter := newReporterProbe()
	claudeFake := process.NewFake()
	codexFake := process.NewFake()
	codexRunning := process.NewFakeRunning()
	codexFake.NextRunning = codexRunning
	rtClaude := startNamedRuntime(t, claudeFake, "rt-claude", "claude")
	rtCodex := startNamedRuntime(t, codexFake, "rt-codex", "codex")

	actor, err := New(Config{
		DaemonID:       "daemon-1",
		Runtimes:       []*runtimeactor.Actor{rtClaude, rtCodex},
		Source:         source,
		Reporter:       reporter,
		Workdir:        workdir.NewFSAdapter(t.TempDir()),
		PollEvery:      10 * time.Millisecond,
		HeartbeatEvery: time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatalf("supervisor Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	seenRegistrations := map[string]bool{}
	for i := 0; i < 2; i++ {
		select {
		case rt := <-source.registered:
			seenRegistrations[rt.RuntimeID] = true
			if rt.Provider != strings.TrimPrefix(rt.RuntimeID, "rt-") {
				t.Fatalf("provider-specific registration mismatch: %+v", rt)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("runtime registration was not published")
		}
	}
	if !seenRegistrations["rt-claude"] || !seenRegistrations["rt-codex"] {
		t.Fatalf("runtime registrations missing: %+v", seenRegistrations)
	}

	select {
	case taskID := <-reporter.started:
		if taskID != "t-codex" {
			t.Fatalf("started task: %q", taskID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("task was not dispatched")
	}
	select {
	case cmd := <-codexRunning.StartedRecv():
		if cmd.Executable != "codex" {
			t.Fatalf("codex runtime command mismatch: %+v", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("codex runtime did not spawn process")
	}

	go func() {
		codexRunning.EmitStdout([]byte("done"))
		codexRunning.EmitExit(0, nil)
	}()
	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("result was not reported")
	}
}

func hasEnvPrefix(env []string, prefix string) bool {
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return true
		}
	}
	return false
}

func readRunEvents(t *testing.T, path string) []ir.CanonicalEvent {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read run event log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	out := make([]ir.CanonicalEvent, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var ev ir.CanonicalEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("decode run event %q: %v", line, err)
		}
		out = append(out, ev)
	}
	return out
}

func assertRunEvent(t *testing.T, events []ir.CanonicalEvent, eventType ir.EventType, check func(ir.CanonicalEvent)) {
	t.Helper()
	for _, ev := range events {
		if ev.Type == eventType {
			if check != nil {
				check(ev)
			}
			return
		}
	}
	t.Fatalf("run event %s not found in %+v", eventType, events)
}

func TestProviderEventDraftMapsCatCEvents(t *testing.T) {
	for _, tc := range []struct {
		name string
		ev   agentbridge.Event
		want ir.EventType
	}{
		{"session", agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: "s-1"}, ir.EventSessionPinned},
		{"text", agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "hello"}, ir.EventTextDelta},
		{"thinking", agentbridge.Event{Kind: agentbridge.EventThinkingDelta, Text: "why"}, ir.EventReasoningDelta},
		{"tool-start", agentbridge.Event{Kind: agentbridge.EventToolCallStarted, Tool: agentbridge.ToolRef{ID: "tool-1", Name: "bash"}}, ir.EventToolCallStarted},
		{"tool-done", agentbridge.Event{Kind: agentbridge.EventToolCallCompleted, Tool: agentbridge.ToolRef{ID: "tool-1", Name: "bash"}}, ir.EventToolCallFinished},
		{"approval", agentbridge.Event{Kind: agentbridge.EventToolApprovalNeeded, Tool: agentbridge.ToolRef{ID: "approval-1", Kind: "exec"}}, ir.EventApprovalRequested},
		{"usage", agentbridge.Event{Kind: agentbridge.EventUsageDelta, Usage: agentbridge.Usage{PromptTokens: 1}}, ir.EventUsageDelta},
		{"warning", agentbridge.Event{Kind: agentbridge.EventWarning, Text: "careful"}, ir.EventLogLine},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, payload, ok := providerEventDraft(tc.ev)
			if !ok {
				t.Fatalf("expected mapping for %+v", tc.ev)
			}
			if got != tc.want {
				t.Fatalf("event type = %s, want %s", got, tc.want)
			}
			if len(payload) == 0 {
				t.Fatalf("payload must not be empty")
			}
		})
	}
	if _, _, ok := providerEventDraft(agentbridge.Event{Kind: agentbridge.EventResult}); ok {
		t.Fatal("EventResult must stay outside non-transition Cat C mapping")
	}
}

func TestProviderEventDraftIncludesToolArgs(t *testing.T) {
	_, payload, ok := providerEventDraft(agentbridge.Event{
		Kind: agentbridge.EventToolCallStarted,
		Tool: agentbridge.ToolRef{
			ID:   "tool-1",
			Name: "Bash",
			Kind: "shell",
			Args: map[string]string{"command": "go test ./..."},
		},
	})
	if !ok {
		t.Fatal("expected tool event mapping")
	}
	args, ok := payload["args"].(map[string]string)
	if !ok {
		t.Fatalf("args payload type = %T", payload["args"])
	}
	if args["command"] != "go test ./..." {
		t.Fatalf("args payload = %+v", args)
	}
}

func TestTerminalResultDraftMapsTaskTransitions(t *testing.T) {
	for _, tc := range []struct {
		name string
		res  agentbridge.Result
		want ir.EventType
	}{
		{"completed", agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}, ir.EventRunReportedDone},
		{"failed", agentbridge.Result{Status: agentbridge.ResultFailed, Error: "boom"}, ir.EventTaskFailed},
		{"blocked", agentbridge.Result{Status: agentbridge.ResultBlocked, Error: "capability"}, ir.EventTaskFailed},
		{"aborted", agentbridge.Result{Status: agentbridge.ResultAborted, Error: "exit"}, ir.EventTaskFailed},
		{"cancelled", agentbridge.Result{Status: agentbridge.ResultCancelled, Error: "user"}, ir.EventTaskCancelled},
		{"timeout", agentbridge.Result{Status: agentbridge.ResultTimeout, Error: "semantic idle timeout"}, ir.EventTaskTimedOut},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, payload := terminalResultDraft(tc.res)
			if got != tc.want {
				t.Fatalf("event type = %s, want %s", got, tc.want)
			}
			if len(payload) == 0 {
				t.Fatalf("payload must not be empty")
			}
			if !got.IsTransition() {
				t.Fatalf("%s must be an IR transition event", got)
			}
		})
	}
}

func TestSupervisorAppliesPolicyBundleToClaudeNativeHooks(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:       "t-claude-hooks",
		Provider: "claude",
		Prompt:   "hello",
		Metadata: map[string]string{
			MetadataWorkspaceID: "ws-1",
		},
	})

	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startNamedRuntime(t, fake, "rt-claude", "claude")
	bundle := policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        "policy-bundle.no-hooks.v1",
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {AllowedSurfaces: policy.AllowedSurfaceSet{}},
		},
	}
	actor, err := New(Config{
		DaemonID:            "daemon-1",
		Runtime:             rt,
		Source:              source,
		Reporter:            reporter,
		Workdir:             workdir.NewFSAdapter(t.TempDir()),
		PollEvery:           10 * time.Millisecond,
		HeartbeatEvery:      time.Hour,
		PolicyBundleVersion: bundle.Version,
		PolicyBundle:        bundle,
		RuntimeTrustTier:    policy.TrustTierHost,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	select {
	case <-reporter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	go func() {
		running.EmitStdout([]byte("ok"))
		running.EmitExit(0, nil)
	}()
	var res agentbridge.Result
	select {
	case res = <-reporter.results:
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("result was not reported")
	}

	manifest := readNativeConfigManifest(t, filepath.Join(res.Workdir, workdir.NativeConfigManifestPath))
	if manifest.ProviderKind != "claude" || manifest.HookMode != workdir.NativeConfigHookModeInstructionOnly {
		t.Fatalf("native config manifest = %+v", manifest)
	}
	for _, blocked := range []string{".claude/settings.json", ".riido/hooks/claude-audit-hook.sh"} {
		if containsString(manifest.GeneratedFiles, blocked) {
			t.Fatalf("manifest generated files must not include %q: %+v", blocked, manifest.GeneratedFiles)
		}
		if _, err := os.Stat(filepath.Join(res.Workdir, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("blocked hook artifact %s should be absent, stat err=%v", blocked, err)
		}
	}
}

func TestSupervisorAppliesPolicyBundleToCodexConfigHome(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:                       "t-codex-home",
		Provider:                 "codex",
		Prompt:                   "hello",
		AllowExperimentalRuntime: true,
		Metadata: map[string]string{
			MetadataWorkspaceID: "ws-1",
		},
	})

	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startNamedRuntime(t, fake, "rt-codex", "codex")
	bundle := policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        "policy-bundle.no-codex-home.v1",
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {AllowedSurfaces: policy.AllowedSurfaceSet{}},
		},
	}
	actor, err := New(Config{
		DaemonID:            "daemon-1",
		Runtime:             rt,
		Source:              source,
		Reporter:            reporter,
		Workdir:             workdir.NewFSAdapter(t.TempDir()),
		PollEvery:           10 * time.Millisecond,
		HeartbeatEvery:      time.Hour,
		PolicyBundleVersion: bundle.Version,
		PolicyBundle:        bundle,
		RuntimeTrustTier:    policy.TrustTierHost,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	select {
	case <-reporter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	go func() {
		running.EmitStdout([]byte("ok"))
		running.EmitExit(0, nil)
	}()
	var res agentbridge.Result
	select {
	case res = <-reporter.results:
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("result was not reported")
	}

	if hasEnvPrefix(running.Command().Env, "TEST_NATIVE_CONFIG_HOME=") {
		t.Fatalf("native config home metadata should be omitted when policy denies it: %+v", running.Command())
	}
	manifest := readNativeConfigManifest(t, filepath.Join(res.Workdir, workdir.NativeConfigManifestPath))
	if manifest.ProviderKind != "codex" || manifest.ConfigHomeDir != "" {
		t.Fatalf("native config manifest = %+v", manifest)
	}
	if containsString(manifest.GeneratedFiles, ".codex/config.toml") {
		t.Fatalf("manifest generated files must not include Codex config home artifact: %+v", manifest.GeneratedFiles)
	}
	if _, err := os.Stat(filepath.Join(res.Workdir, ".codex", "config.toml")); !os.IsNotExist(err) {
		t.Fatalf("Codex config home artifact should be absent, stat err=%v", err)
	}
}

func TestSupervisorKeepsOpenClawAndCursorNativeConfigInstructionOnly(t *testing.T) {
	for _, provider := range []bridge.Provider{"openclaw", "cursor"} {
		t.Run(string(provider), func(t *testing.T) {
			source := controlplane.NewMemorySource()
			source.Enqueue(bridge.TaskRequest{
				ID:                       "t-" + string(provider) + "-native-config",
				Provider:                 provider,
				Prompt:                   "hello",
				AllowExperimentalRuntime: true,
				Metadata: map[string]string{
					MetadataWorkspaceID: "ws-1",
				},
			})

			reporter := newReporterProbe()
			fake := process.NewFake()
			running := process.NewFakeRunning()
			fake.NextRunning = running
			rt := startNamedRuntime(t, fake, "rt-"+string(provider), string(provider))
			actor, err := New(Config{
				DaemonID:            "daemon-1",
				Runtime:             rt,
				Source:              source,
				Reporter:            reporter,
				Workdir:             workdir.NewFSAdapter(t.TempDir()),
				PollEvery:           10 * time.Millisecond,
				HeartbeatEvery:      time.Hour,
				PolicyBundleVersion: policy.DefaultLocalPolicyBundleVersion,
				PolicyBundle:        policy.DefaultLocalPolicyBundle(),
				RuntimeTrustTier:    policy.TrustTierHost,
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := actor.Start(context.Background()); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_ = actor.Stop(ctx)
			})

			select {
			case <-reporter.started:
			case <-time.After(2 * time.Second):
				t.Fatal("task was not claimed")
			}

			select {
			case <-running.StartedRecv():
			case <-time.After(2 * time.Second):
				t.Fatal("provider process was not spawned")
			}

			go func() {
				running.EmitStdout([]byte("ok"))
				running.EmitExit(0, nil)
			}()
			var res agentbridge.Result
			select {
			case res = <-reporter.results:
				if res.Status != agentbridge.ResultCompleted {
					t.Fatalf("result: %+v", res)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("result was not reported")
			}

			if hasEnvPrefix(running.Command().Env, "TEST_NATIVE_CONFIG_HOME=") {
				t.Fatalf("native config home metadata must be omitted for %s: %+v", provider, running.Command())
			}
			manifest := readNativeConfigManifest(t, filepath.Join(res.Workdir, workdir.NativeConfigManifestPath))
			if manifest.ProviderKind != string(provider) ||
				manifest.PrimaryInstructionFile != "AGENTS.md" ||
				manifest.HookMode != workdir.NativeConfigHookModeInstructionOnly ||
				manifest.ConfigHomeDir != "" ||
				len(manifest.ProviderSettingsFiles) != 0 ||
				len(manifest.HookFiles) != 0 {
				t.Fatalf("native config manifest = %+v", manifest)
			}
			if len(manifest.GeneratedFiles) != 2 ||
				!containsString(manifest.GeneratedFiles, "AGENTS.md") ||
				!containsString(manifest.GeneratedFiles, workdir.NativeConfigManifestPath) {
				t.Fatalf("generated files = %+v", manifest.GeneratedFiles)
			}
			for _, blocked := range []string{
				".cursor/settings.json",
				".cursor/rules",
				".openclaw/settings.json",
				".openclaw/config.json",
			} {
				if _, err := os.Stat(filepath.Join(res.Workdir, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
					t.Fatalf("provider-native artifact %s should be absent, stat err=%v", blocked, err)
				}
			}
		})
	}
}

func TestSupervisorStopArchivesInFlightWorkspace(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:       "t-stop",
		Provider: "fake",
		Prompt:   "x",
		Metadata: map[string]string{
			MetadataWorkspaceID: "ws-stop",
			MetadataRunID:       "run-stop",
		},
	})

	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)

	actor, err := New(Config{
		DaemonID:           "daemon-stop",
		RiidoDaemonVersion: "riido-agentd v1.2.3",
		Runtime:            rt,
		Source:             source,
		Reporter:           reporter,
		Workdir:            workdir.NewFSAdapter(t.TempDir()),
		PollEvery:          10 * time.Millisecond,
		HeartbeatEvery:     time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	select {
	case <-reporter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	var cmd process.Command
	select {
	case cmd = <-running.StartedRecv():
		if cmd.Dir == "" {
			t.Fatalf("provider command missing workdir: %+v", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("provider process was not spawned")
	}

	running.EmitStdout([]byte("event"))
	select {
	case ev := <-reporter.events:
		if ev.Kind != agentbridge.EventLifecycle || ev.Phase != agentbridge.StateRunning {
			t.Fatalf("running event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("running event was not reported")
	}

	select {
	case ev := <-reporter.events:
		if ev.Kind != agentbridge.EventTextDelta || ev.Text != "event" {
			t.Fatalf("event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("nonterminal event was not reported")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := actor.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	select {
	case <-running.KillRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("provider process was not killed on supervisor stop")
	}

	var res agentbridge.Result
	select {
	case res = <-reporter.results:
		if res.Status != agentbridge.ResultCancelled || !strings.Contains(res.Error, "supervisor: stopped") {
			t.Fatalf("shutdown result: %+v", res)
		}
		if res.Workdir != cmd.Dir {
			t.Fatalf("shutdown result workdir = %q, want %q", res.Workdir, cmd.Dir)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown result was not reported")
	}

	runRoot := filepath.Dir(res.Workdir)
	archive, err := os.ReadFile(filepath.Join(runRoot, "archive.json"))
	if err != nil {
		t.Fatalf("archive manifest not written on stop: %v", err)
	}
	if !strings.Contains(string(archive), `"result_status": "cancelled"`) {
		t.Fatalf("archive manifest should record cancelled status:\n%s", archive)
	}

	events := readRunEvents(t, filepath.Join(runRoot, "ir", "events.jsonl"))
	assertRunEvent(t, events, ir.EventTaskCancelled, func(ev ir.CanonicalEvent) {
		if ev.ActorKind != ir.ActorDaemon {
			t.Fatalf("TaskCancelled must be daemon-attributed: %+v", ev)
		}
		if ev.FSMVersion != task.FSMSchemaVersion {
			t.Fatalf("TaskCancelled FSMVersion = %d, want %d", ev.FSMVersion, task.FSMSchemaVersion)
		}
	})
	assertRunEvent(t, events, ir.EventWorkdirArchived, nil)
}

func TestSupervisorWorkdirRequiresWorkspaceID(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{ID: "t-no-workspace", Provider: "fake", Prompt: "x"})

	reporter := newReporterProbe()
	rt := startRuntime(t, process.NewFake())
	actor, err := New(Config{
		DaemonID:       "daemon-1",
		Runtime:        rt,
		Source:         source,
		Reporter:       reporter,
		Workdir:        workdir.NewFSAdapter(t.TempDir()),
		PollEvery:      10 * time.Millisecond,
		HeartbeatEvery: time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	select {
	case <-reporter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultFailed || res.Error == "" {
			t.Fatalf("expected workdir failure result, got %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("workdir failure was not reported")
	}
}

func TestSupervisorBlocksTaskWhenRequiredSurfaceUnsupported(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:               "t-needs-mcp",
		Provider:         "fake",
		Prompt:           "x",
		RequiredSurfaces: []string{"mcp"},
	})

	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)
	actor, err := New(Config{
		DaemonID:       "daemon-1",
		Runtime:        rt,
		Source:         source,
		Reporter:       reporter,
		PollEvery:      10 * time.Millisecond,
		HeartbeatEvery: time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	select {
	case <-reporter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultBlocked {
			t.Fatalf("expected blocked result, got %+v", res)
		}
		if !strings.Contains(res.Error, "MISSING_REQUIRED_SURFACE:mcp") {
			t.Fatalf("missing scheduler reason: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ineligible task was not reported")
	}
	if running.Command().Executable != "" {
		t.Fatalf("provider process should not have spawned: %+v", running.Command())
	}
}

type cancelSource struct {
	req    bridge.TaskRequest
	cancel chan error
}

func (s *cancelSource) RegisterRuntime(context.Context, controlplane.RuntimeRegistration) error {
	return nil
}
func (s *cancelSource) DeregisterRuntime(context.Context, string) error { return nil }
func (s *cancelSource) Heartbeat(context.Context, controlplane.RuntimeHeartbeat) error {
	return nil
}
func (s *cancelSource) ClaimTask(context.Context, string) (*bridge.TaskRequest, error) {
	if s.req.ID == "" {
		return nil, nil
	}
	req := s.req
	s.req = bridge.TaskRequest{}
	return &req, nil
}
func (s *cancelSource) WatchCancellation(context.Context, string) (<-chan error, error) {
	return s.cancel, nil
}

func TestSupervisorRoutesCancellationToRuntime(t *testing.T) {
	source := &cancelSource{
		req:    bridge.TaskRequest{ID: "t-cancel", Provider: "fake", Prompt: "x"},
		cancel: make(chan error, 1),
	}
	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)

	actor, err := New(Config{
		DaemonID:       "daemon-1",
		Runtime:        rt,
		Source:         source,
		Reporter:       reporter,
		PollEvery:      10 * time.Millisecond,
		HeartbeatEvery: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	select {
	case <-reporter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	source.cancel <- errors.New("human cancel")

	select {
	case <-running.KillRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("provider process was not killed")
	}

	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCancelled {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancel result was not reported")
	}
}
