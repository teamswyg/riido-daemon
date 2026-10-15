package supervisor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/execution"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type correctionAdapter struct {
	stubAdapter
	requests chan agentbridge.StartRequest
}

func (a *correctionAdapter) Detect(context.Context, agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true, Executable: "fake", SupportsResume: true}, nil
}

func (a *correctionAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	a.requests <- req
	return agentbridge.StartCommand{Executable: "fake", Dir: req.Cwd}, nil
}

func (a *correctionAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	events, commands, err := a.stubAdapter.Translate(raw)
	events = append([]agentbridge.Event{{Kind: agentbridge.EventSessionIdentified, SessionID: "native-session"}}, events...)
	for i := range events {
		if events[i].Kind == agentbridge.EventResult {
			events[i].Result.SessionID = "native-session"
		}
	}
	return events, commands, err
}

type correctionProcess struct{ started chan *process.FakeRunning }

func (p *correctionProcess) Start(ctx context.Context, cmd process.Command) (process.RunningProcess, error) {
	r := process.NewFakeRunning()
	if _, err := (&process.Fake{NextRunning: r}).Start(ctx, cmd); err != nil {
		return nil, err
	}
	p.started <- r
	return r, nil
}

func correctionRepository(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "result.txt"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-q", repo}, {"-C", repo, "add", "."}, {"-C", repo, "-c", "user.name=test", "-c", "user.email=test@example.invalid", "commit", "-qm", "fixture"}} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("%s: %v", out, err)
		}
	}
	return repo
}

func startCorrectionActor(t *testing.T, limit int, checks ...[]string) (*correctionAdapter, *correctionProcess, *reporterProbe, *controlplane.MemorySource, string) {
	t.Helper()
	repo := correctionRepository(t)
	if len(checks) == 0 {
		checks = [][]string{{"git", "diff", "--check"}}
	}
	adapter := &correctionAdapter{stubAdapter: stubAdapter{name: "fake"}, requests: make(chan agentbridge.StartRequest, 4)}
	proc := &correctionProcess{started: make(chan *process.FakeRunning, 4)}
	rt, err := runtimeactor.New(runtimeactor.Config{RuntimeID: "rt", MaxConcurrent: 2, Adapters: []agentbridge.Adapter{adapter}, Process: proc})
	if err != nil {
		t.Fatal(err)
	}
	if err = rt.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	source := controlplane.NewMemorySource()
	reporter := newReporterProbe()
	reporter.events = make(chan agentbridge.Event, 1000)
	source.Enqueue(bridge.TaskRequest{ID: "assignment", Provider: "fake", Prompt: "Fix result", Model: "selected-model", Metadata: map[string]string{"workspace_id": "workspace", "task_id": "task"}})
	a, err := New(Config{DaemonID: "daemon", RiidoDaemonVersion: "test-v1", Runtime: rt, Source: source, Reporter: reporter, Workdir: workdir.NewFSAdapter(t.TempDir()), PollEvery: time.Millisecond, HeartbeatEvery: 20 * time.Millisecond, ExecutionBindings: []execution.Binding{{WorkspaceID: "workspace", TaskID: "task", Repository: repo, ChangedFiles: []string{"result.txt"}, Checks: checks, ContinuationLimit: limit}}})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = a.Stop(ctx)
		_ = rt.Stop(ctx)
	})
	return adapter, proc, reporter, source, repo
}

func awaitCorrectionRequest(t *testing.T, adapter *correctionAdapter, proc *correctionProcess) (agentbridge.StartRequest, *process.FakeRunning) {
	t.Helper()
	select {
	case req := <-adapter.requests:
		select {
		case p := <-proc.started:
			return req, p
		case <-time.After(3 * time.Second):
			t.Fatal("provider not spawned")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("provider request missing")
	}
	return agentbridge.StartRequest{}, nil
}
