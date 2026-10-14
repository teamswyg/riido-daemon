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
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/execution"
)

func TestExecutionCorrectionResumesSameWorkspaceAndAssignment(t *testing.T) {
	adapter, proc, reporter, _, sourceRepo := startCorrectionActor(t, 1)
	first, p := awaitCorrectionRequest(t, adapter, proc)
	if err := os.WriteFile(filepath.Join(first.Cwd, "preserve.txt"), []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	p.EmitStdout([]byte("analysis only"))
	next, p2 := awaitCorrectionRequest(t, adapter, proc)
	if next.TaskID != first.TaskID || next.Cwd != first.Cwd || next.Model != first.Model || next.ResumeSessionID != "native-session" || !strings.Contains(next.Prompt, "not accept completion") {
		t.Fatalf("correction lost execution identity: first=%+v next=%+v", first, next)
	}
	if b, err := os.ReadFile(filepath.Join(next.Cwd, "preserve.txt")); err != nil || string(b) != "keep" {
		t.Fatalf("existing work overwritten: %s %v", b, err)
	}
	if err := os.WriteFile(filepath.Join(next.Cwd, "result.txt"), []byte("fixed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	p2.EmitStdout([]byte("fixed"))
	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCompleted || !strings.Contains(res.Output, "checks passed: 1") {
			t.Fatalf("completion unverified: %+v", res)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no terminal result")
	}
	if b, _ := os.ReadFile(filepath.Join(sourceRepo, "result.txt")); string(b) != "old" {
		t.Fatal("source repository modified")
	}
	if len(reporter.started) != 1 {
		t.Fatalf("assignment reclaimed: %d starts", len(reporter.started))
	}
}

func TestExecutionCorrectionStopsAtHostLimit(t *testing.T) {
	adapter, proc, reporter, _, _ := startCorrectionActor(t, 1)
	_, p := awaitCorrectionRequest(t, adapter, proc)
	p.EmitStdout([]byte("analysis only"))
	_, p2 := awaitCorrectionRequest(t, adapter, proc)
	p2.EmitStdout([]byte("still analysis"))
	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultFailed || !strings.Contains(res.Error, "not changed") {
			t.Fatalf("missing work accepted: %+v", res)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("correction did not terminate")
	}
	if len(adapter.requests) != 0 {
		t.Fatal("unbounded provider replay")
	}
}

func TestExecutionCorrectionRefusesUnsafeRecovery(t *testing.T) {
	for _, kind := range []string{"invalid_output", "verification_timeout", "evidence_persistence", ""} {
		task := &runningTask{ctx: context.Background(), execution: &execution.Prepared{Binding: execution.Binding{ContinuationLimit: 1}}, canResume: true, request: &bridge.TaskRequest{}, runtime: &runtimeactor.Actor{}}
		if canContinueExecution(task, agentbridge.Result{Status: agentbridge.ResultFailed, SessionID: "native"}, execution.Evidence{ErrorKind: kind}) {
			t.Fatalf("unsafe failure resumed: %s", kind)
		}
	}
}

func TestExecutionCorrectionHonorsCancellation(t *testing.T) {
	adapter, proc, reporter, source, _ := startCorrectionActor(t, 1)
	_, first := awaitCorrectionRequest(t, adapter, proc)
	first.EmitStdout([]byte("analysis only"))
	_, _ = awaitCorrectionRequest(t, adapter, proc)
	source.Cancel("assignment", context.Canceled)
	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCancelled {
			t.Fatalf("cancelled correction completed: %+v", res)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("correction cancellation blocked")
	}
	if len(adapter.requests) != 0 {
		t.Fatal("cancelled task replayed")
	}
}
