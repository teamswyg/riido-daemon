package supervisor

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

// Explicit opt-in: uses the installed, authenticated Codex CLI and consumes a turn.
func TestIntegrationAnswerOnlyUsesOneProviderTurn(t *testing.T) {
	if os.Getenv("AGENTBRIDGE_INTEGRATION") != "1" {
		t.Skip("AGENTBRIDGE_INTEGRATION not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	adapter := &singlePassAdapter{nativeCorrectionAdapter: nativeCorrectionAdapter{requests: make(chan agentbridge.StartRequest, 8), names: make(chan string, 8)}}
	rt, err := runtimeactor.New(runtimeactor.Config{RuntimeID: "single-pass-codex", Adapters: []agentbridge.Adapter{adapter}, Process: processexec.New(), HardTimeout: 60 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if err = rt.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() {
		c, stop := context.WithTimeout(context.Background(), 3*time.Second)
		defer stop()
		_ = rt.Stop(c)
	}()
	source := controlplane.NewMemorySource()
	reporter := newReporterProbe()
	reporter.events = make(chan agentbridge.Event, 4096)
	prompt, system, _, _ := agentbridge.ApplyRuntimeInstructionContract("codex", "Reply with exactly RIIDO_SINGLE_PASS_OK. Do not use tools, read files, or create a repository.", "", "")
	source.Enqueue(bridge.TaskRequest{ID: "single-pass", Provider: "codex", AllowExperimentalRuntime: true, Model: os.Getenv("AGENTBRIDGE_CODEX_MODEL"), Prompt: prompt, SystemPrompt: system, Metadata: map[string]string{"workspace_id": "workspace", "task_id": "task"}})
	a, err := New(Config{DaemonID: "local-test", RiidoDaemonVersion: "test", Runtime: rt, Source: source, Reporter: reporter, Workdir: workdir.NewFSAdapter(t.TempDir()), PollEvery: 10 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() {
		c, stop := context.WithTimeout(context.Background(), 3*time.Second)
		defer stop()
		_ = a.Stop(c)
	}()
	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCompleted || !strings.Contains(res.Output, "RIIDO_SINGLE_PASS_OK") {
			t.Fatalf("status=%s error=%s output=%s", res.Status, res.Error, res.Output)
		}
		if adapter.turns.Load() != 1 || len(adapter.requests) != 1 {
			t.Fatalf("turn/start=%d process starts=%d", adapter.turns.Load(), len(adapter.requests))
		}
		if res.ExecutionContext != nil && res.ExecutionContext.VerificationStatus != "not_configured" {
			t.Fatalf("answer-only run required host checks: %+v", res.ExecutionContext)
		}
		t.Logf("turn/start=1 process_starts=1 provider_elapsed_ms=%d", res.FinishedAt.Sub(res.StartedAt).Milliseconds())
	case <-ctx.Done():
		t.Fatal("single-pass response timed out")
	}
}

type singlePassAdapter struct {
	nativeCorrectionAdapter
	turns atomic.Int32
}

func (a *singlePassAdapter) NewProtocolDriver(req agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	d, err := a.nativeCorrectionAdapter.NewProtocolDriver(req)
	return &singlePassDriver{ProtocolDriver: d, turns: &a.turns}, err
}

type singlePassDriver struct {
	agentbridge.ProtocolDriver
	turns *atomic.Int32
}

func (d *singlePassDriver) OnStart(ctx context.Context, io agentbridge.ProtocolIO) error {
	return d.ProtocolDriver.OnStart(ctx, singlePassIO{io, d.turns})
}

func (d *singlePassDriver) OnRaw(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	return d.ProtocolDriver.OnRaw(ctx, raw, singlePassIO{io, d.turns})
}

type singlePassIO struct {
	agentbridge.ProtocolIO
	turns *atomic.Int32
}

func (io singlePassIO) WriteStdin(ctx context.Context, b []byte) error {
	var frame struct {
		Method string `json:"method"`
	}
	if json.Unmarshal(b, &frame) == nil && frame.Method == "turn/start" {
		io.turns.Add(1)
	}
	return io.ProtocolIO.WriteStdin(ctx, b)
}
