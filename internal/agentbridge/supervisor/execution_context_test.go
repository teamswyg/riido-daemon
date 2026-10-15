package supervisor

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/execution"
)

func TestReportedCompletionDoesNotInventHostVerification(t *testing.T) {
	got := resultExecutionContext(&runningTask{}, agentbridge.Result{Status: agentbridge.ResultCompleted})
	if got.VerificationStatus != "not_configured" {
		t.Fatal("provider reply was presented as host verified")
	}
	task := &runningTask{execution: &execution.Prepared{Binding: execution.Binding{ChangedFiles: []string{"out"}, Checks: [][]string{{"git", "diff", "--check"}}}}, executionVerified: true}
	got = resultExecutionContext(task, agentbridge.Result{Status: agentbridge.ResultCompleted})
	if got.VerificationStatus != "passed" || got.WorkspaceKind != "repository_clone" {
		t.Fatal("verified host execution lost evidence")
	}
}

func TestExecutionFailuresKeepTypedCauseAndCancellationPrecedence(t *testing.T) {
	for _, row := range []struct {
		err  error
		code string
	}{
		{errors.Join(errWorkspacePreparation, errors.New("private path")), "workspace_preparation_failed"},
		{errors.Join(errWorkspacePreparation, errNativeSessionBinding), "session_scope_conflict"},
		{context.Canceled, "run_cancelled"},
	} {
		result := resultForActivationError(row.err)
		context := resultExecutionContext(nil, result)
		if context.DiagnosticCode != row.code {
			t.Fatalf("got %s, want %s", context.DiagnosticCode, row.code)
		}
	}
	task := &runningTask{execution: &execution.Prepared{Binding: execution.Binding{ChangedFiles: []string{"out"}, Checks: [][]string{{"git", "diff", "--check"}}}}, verificationPending: true}
	got := resultExecutionContext(task, agentbridge.Result{Status: agentbridge.ResultCancelled})
	if got.VerificationStatus != "cancelled" || got.DiagnosticCode != "run_cancelled" {
		t.Fatal("cancelled verification presented as passed or retryable")
	}
}

func TestFailedExecutionDoesNotLeaveHostChecksPending(t *testing.T) {
	task := &runningTask{execution: &execution.Prepared{Binding: execution.Binding{ChangedFiles: []string{"out"}, Checks: [][]string{{"git", "diff", "--check"}}}}}
	got := resultExecutionContext(task, agentbridge.Result{Status: agentbridge.ResultTimeout})
	if got.VerificationStatus != "not_run" {
		t.Fatalf("checks never started: %s", got.VerificationStatus)
	}
}

func TestBindingSelectionIsFrozenForCurrentExecution(t *testing.T) {
	store, err := execution.OpenBindingStore(filepath.Join(t.TempDir(), "bindings.json"))
	if err != nil {
		t.Fatal(err)
	}
	binding := execution.Binding{WorkspaceID: "w", TaskID: "t", Repository: t.TempDir(), ChangedFiles: []string{"out"}, Checks: [][]string{{"git", "diff", "--check"}}}
	if err := store.Replace("w", "t", &binding); err != nil {
		t.Fatal(err)
	}
	actor := &Actor{cfg: Config{ExecutionBindingStore: store}}
	request := &bridge.TaskRequest{HostExecutionBindingSelected: true, HostExecutionBinding: &binding}
	replacement := binding
	replacement.Repository = t.TempDir()
	if err := store.Replace("w", "t", &replacement); err != nil {
		t.Fatal(err)
	}
	got := actor.nativeSessionExecutionBinding(request)
	if got.Repository != binding.Repository {
		t.Fatal("running execution switched repository")
	}
	request.HostExecutionBinding = nil
	if actor.nativeSessionExecutionBinding(request) != nil {
		t.Fatal("new configuration retroactively bound an isolated run")
	}
}

func TestRepositorySelectionWithoutChecksIsNotVerified(t *testing.T) {
	task := &runningTask{execution: &execution.Prepared{Binding: execution.Binding{Repository: t.TempDir()}}, executionVerified: true}
	got := resultExecutionContext(task, agentbridge.Result{Status: agentbridge.ResultCompleted})
	if got.WorkspaceKind != "repository_clone" || got.VerificationStatus != "not_configured" {
		t.Fatalf("folder selection invented host evidence: %+v", got)
	}
}
