package supervisor

import (
	"context"
	"encoding/json"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
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
	return slices.Contains(values, want)
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

type lifecycleReporterProbe struct {
	*reporterProbe
	completeLevels chan lifecycle.ShutdownLevel
}

func newLifecycleReporterProbe() *lifecycleReporterProbe {
	return &lifecycleReporterProbe{
		reporterProbe:  newReporterProbe(),
		completeLevels: make(chan lifecycle.ShutdownLevel, 4),
	}
}

func (r *lifecycleReporterProbe) CompleteTask(ctx context.Context, taskID string, res agentbridge.Result) error {
	r.completeLevels <- lifecycle.FromContext(ctx).ShutdownLevel()
	return r.reporterProbe.CompleteTask(ctx, taskID, res)
}

func startRuntime(t *testing.T, fake *process.Fake) *runtimeactor.Actor {
	t.Helper()
	return startNamedRuntime(t, fake, "rt-local", "fake")
}

func startNamedRuntime(t *testing.T, fake *process.Fake, runtimeID, provider string) *runtimeactor.Actor {
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
