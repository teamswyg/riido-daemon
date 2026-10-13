package supervisor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/execution"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestIntegrationExecutionCorrection(t *testing.T) {
	if os.Getenv("AGENTBRIDGE_INTEGRATION") != "1" {
		t.Skip("AGENTBRIDGE_INTEGRATION not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	repo := correctionRepository(t)
	adapter := &nativeCorrectionAdapter{requests: make(chan agentbridge.StartRequest, 4), names: make(chan string, 4)}
	rt, err := runtimeactor.New(runtimeactor.Config{RuntimeID: "native-codex", Adapters: []agentbridge.Adapter{adapter}, Process: processexec.New(), HardTimeout: 60 * time.Second})
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
	reporter.events = make(chan agentbridge.Event, 1024)
	prompt, system, _, _ := agentbridge.ApplyRuntimeInstructionContract("codex", "Inspect result.txt and report that it contains old. Do not modify it in this turn. The overall assigned goal is to replace its content with fixed in the next turn, after host verification feedback.", "", "Include RIIDO_NATIVE_CORRECTION_ACK in your final reply.")
	source.Enqueue(bridge.TaskRequest{ID: "native-assignment", Provider: "codex", AllowExperimentalRuntime: true, Model: os.Getenv("AGENTBRIDGE_CODEX_MODEL"), Prompt: prompt, SystemPrompt: system, Metadata: map[string]string{"workspace_id": "workspace", "task_id": "task", "task_title": "Agent 작업 수행 개선 · native 검증"}})
	a, err := New(Config{DaemonID: "local-test", RiidoDaemonVersion: "test", Runtime: rt, Source: source, Reporter: reporter, Workdir: workdir.NewFSAdapter(t.TempDir()), PollEvery: 10 * time.Millisecond, ExecutionBindings: []execution.Binding{{WorkspaceID: "workspace", TaskID: "task", Repository: repo, ChangedFiles: []string{"result.txt"}, Checks: [][]string{{"git", "diff", "--check"}}, ContinuationLimit: 1}}})
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
		if res.Status != agentbridge.ResultCompleted || !strings.Contains(res.Output, "RIIDO_NATIVE_CORRECTION_ACK") {
			t.Fatalf("native correction failed: status=%s error=%s output=%s", res.Status, res.Error, res.Output)
		}
		if len(adapter.requests) != 2 {
			t.Fatalf("expected native correction turn; observed %d starts", len(adapter.requests))
		}
		first, next := <-adapter.requests, <-adapter.requests
		if first.Cwd != next.Cwd || next.ResumeSessionID == "" || first.TaskID != next.TaskID {
			t.Fatal("native correction lost identity")
		}
		body, err := os.ReadFile(filepath.Join(first.Cwd, "result.txt"))
		if err != nil || strings.TrimSpace(string(body)) != "fixed" {
			t.Fatalf("actual correction missing: %s %v", body, err)
		}
		original, _ := os.ReadFile(filepath.Join(repo, "result.txt"))
		if string(original) != "old" {
			t.Fatal("source repository modified")
		}
		select {
		case name := <-adapter.names:
			if name != "Agent 작업 수행 개선 · native 검증" {
				t.Fatalf("native task name=%q", name)
			}
		default:
			t.Fatal("native task name update not observed")
		}
		t.Log("real native resume, same clone, developer instruction marker, file change and host check passed")
	case <-ctx.Done():
		t.Fatal("native correction timed out")
	}
}
