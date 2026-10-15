package supervisor

import (
	"context"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type recoveryAdapter struct {
	stubAdapter
	unavailable hostintegration.ProviderDiagnosticCode
	requests    chan agentbridge.StartRequest
}

func (a *recoveryAdapter) Detect(context.Context, agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	if a.unavailable != "" {
		return agentbridge.DetectResult{Available: false, DiagnosticCode: a.unavailable, HealthStatus: hostintegration.ProviderHealthUnavailable}, nil
	}
	return agentbridge.DetectResult{Available: true, Executable: a.name, Version: "1.0", HealthStatus: hostintegration.ProviderHealthHealthy, DiagnosticCode: hostintegration.ProviderDiagnosticNone, SupportsStreaming: true, SupportsResume: true, SupportsSystem: true, SupportsMaxTurns: true, SupportsMCP: true, SupportsToolHooks: true, SupportsUsage: true}, nil
}

func (a *recoveryAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	a.requests <- req
	return agentbridge.StartCommand{Executable: a.name, Dir: req.Cwd}, nil
}

type retryStartProcess struct {
	mu       sync.Mutex
	failures int
	calls    int
	dirs     []string
	started  chan *process.FakeRunning
}

func (p *retryStartProcess) Start(ctx context.Context, cmd process.Command) (process.RunningProcess, error) {
	p.mu.Lock()
	p.calls++
	p.dirs = append(p.dirs, cmd.Dir)
	fail := p.calls <= p.failures
	p.mu.Unlock()
	if fail {
		return nil, syscall.EAGAIN
	}
	run := process.NewFakeRunning()
	_, err := (&process.Fake{NextRunning: run}).Start(ctx, cmd)
	if err != nil {
		return nil, err
	}
	p.started <- run
	return run, nil
}
func (p *retryStartProcess) count() int { p.mu.Lock(); defer p.mu.Unlock(); return p.calls }
func newRecoveryRuntime(t *testing.T, name string, unavailable hostintegration.ProviderDiagnosticCode, p *retryStartProcess, modelIDs ...string) (*runtimeactor.Actor, *recoveryAdapter) {
	t.Helper()
	adapter := &recoveryAdapter{stubAdapter: stubAdapter{name: name}, unavailable: unavailable, requests: make(chan agentbridge.StartRequest, 20)}
	if len(modelIDs) == 0 {
		modelIDs = []string{"chosen-model"}
	}
	var models []runtimeactor.RuntimeModel
	for _, id := range modelIDs {
		models = append(models, runtimeactor.RuntimeModel{ModelID: id})
	}
	rt, err := runtimeactor.New(runtimeactor.Config{RuntimeID: "rt:" + name, Adapters: []agentbridge.Adapter{adapter}, Process: p, Models: models})
	if err != nil {
		t.Fatal(err)
	}
	if err = rt.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = rt.Stop(ctx)
	})
	return rt, adapter
}

func startRecoverySupervisor(t *testing.T, req bridge.TaskRequest, runtimes ...*runtimeactor.Actor) (*Actor, *reporterProbe, *controlplane.MemorySource) {
	t.Helper()
	source := controlplane.NewMemorySource()
	source.Enqueue(req)
	reporter := newReporterProbe()
	reporter.events = make(chan agentbridge.Event, 100)
	actor := startRoutingSupervisor(t, Config{DaemonID: "daemon", Source: source, Reporter: reporter, Runtimes: runtimes, Workdir: workdir.NewFSAdapter(t.TempDir())})
	return actor, reporter, source
}

func recoveryRequest() bridge.TaskRequest {
	return bridge.TaskRequest{ID: "recover-execution", Provider: "fake", Prompt: "Fix the file", ProviderNeutralPrompt: "Fix the file", Metadata: map[string]string{"workspace_id": "workspace", "task_id": "task"}}
}

func awaitRecoverySpawn(t *testing.T, p *retryStartProcess) *process.FakeRunning {
	t.Helper()
	select {
	case run := <-p.started:
		return run
	case <-time.After(4 * time.Second):
		t.Fatal("recovered provider did not start")
	}
	return nil
}
