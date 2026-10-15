package supervisor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/execution"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func historyRequest(run, resume string) *bridge.TaskRequest {
	return &bridge.TaskRequest{ID: run, Provider: "claude", ResumeSessionID: resume, Metadata: map[string]string{"workspace_id": "workspace", "task_id": "task", "agent_id": "agent"}}
}

func historyActor(root string) (*Actor, runtimeactor.Status) {
	capView := runtimeactor.Capability{Provider: "claude", Available: true, SupportsResume: true, CapabilityFingerprint: "runtime-pin", ProtocolKind: "claude-stream-json"}
	return &Actor{cfg: Config{Workdir: workdir.NewFSAdapter(root), DaemonID: "test", RiidoDaemonVersion: "test", PolicyBundleVersion: "test-policy"}, mailbox: make(chan envelope, 20), stoppedCh: make(chan struct{})}, runtimeactor.Status{RuntimeID: "runtime", Capabilities: []runtimeactor.Capability{capView}}
}

func prepareAndPinHistory(t *testing.T, a *Actor, status runtimeactor.Status, req *bridge.TaskRequest) *preparedWorkspace {
	t.Helper()
	prepared, err := a.prepareWorkspace(context.Background(), status, req)
	if err != nil {
		t.Fatal(err)
	}
	task := &runningTask{request: req, workspace: prepared.workspace, events: prepared.events}
	if err := a.pinTaskNativeSession(task, agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: "session"}); err != nil {
		t.Fatal(err)
	}
	return prepared
}

func TestNativeSessionFollowupsRetainWorkspaceAcrossRestart(t *testing.T) {
	root := t.TempDir()
	a, status := historyActor(root)
	first := prepareAndPinHistory(t, a, status, historyRequest("first", ""))
	marker := filepath.Join(first.workspace.Workdir, "result.txt")
	if err := os.WriteFile(marker, []byte("first turn"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, run := range []string{"second", "third"} {
		a, status = historyActor(root)
		next := prepareAndPinHistory(t, a, status, historyRequest(run, "session"))
		if next.workspace.Workdir != first.workspace.Workdir || next.workspace.Root == first.workspace.Root || next.workspace.IR == first.workspace.IR {
			t.Fatalf("workspace continuity: %+v %+v", first.workspace, next.workspace)
		}
		if body, err := os.ReadFile(marker); err != nil || string(body) != "first turn" {
			t.Fatalf("retained work: %s %v", body, err)
		}
	}
}

func TestNativeSessionResumeRejectsChangedOwnerRuntimeAndGrant(t *testing.T) {
	for _, field := range []string{"workspace_id", "task_id", "agent_id", "runtime", "grant"} {
		t.Run(field, func(t *testing.T) {
			a, status := historyActor(t.TempDir())
			prepareAndPinHistory(t, a, status, historyRequest("first", ""))
			req := historyRequest("next", "session")
			switch field {
			case "runtime":
				status.Capabilities[0].CapabilityFingerprint = "other"
			case "grant":
				a.cfg.ExecutionBindings = []execution.Binding{{WorkspaceID: "workspace", TaskID: "task", Repository: "/another/repository"}}
			default:
				req.Metadata[field] = "another"
			}
			if _, err := a.prepareWorkspace(context.Background(), status, req); !errors.Is(err, errNativeSessionBinding) {
				t.Fatalf("expected blocked resume, got %v", err)
			}
		})
	}
}

func TestNativeSessionResumeDoesNotRecreateMissingHistory(t *testing.T) {
	a, status := historyActor(t.TempDir())
	first := prepareAndPinHistory(t, a, status, historyRequest("same", ""))
	if err := os.RemoveAll(first.workspace.Root); err != nil {
		t.Fatal(err)
	}
	if _, err := a.prepareWorkspace(context.Background(), status, historyRequest("same", "session")); !errors.Is(err, errNativeSessionBinding) {
		t.Fatalf("missing history: %v", err)
	}
	if _, err := os.Stat(first.workspace.Root); !os.IsNotExist(err) {
		t.Fatalf("history was silently recreated: %v", err)
	}
}

func TestNativeSessionFollowupReopensCloneWithFreshVerification(t *testing.T) {
	a, status := historyActor(t.TempDir())
	a.cfg.ExecutionBindings = []execution.Binding{{WorkspaceID: "workspace", TaskID: "task", Repository: correctionRepository(t), ChangedFiles: []string{"result.txt"}, Checks: [][]string{{"git", "diff", "--check"}}}}
	first := prepareAndPinHistory(t, a, status, historyRequest("first", ""))
	file := filepath.Join(first.execution.Directory, "result.txt")
	if err := os.WriteFile(file, []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	if evidence := first.execution.Verify(context.Background()); evidence.Error != "" {
		t.Fatal(evidence.Error)
	}
	next := prepareAndPinHistory(t, a, status, historyRequest("next", "session"))
	if next.execution.Directory != first.execution.Directory {
		t.Fatal("clone changed")
	}
	if evidence := next.execution.Verify(context.Background()); evidence.ErrorKind != "unchanged_output" {
		t.Fatalf("old work accepted for new turn: %+v", evidence)
	}
	if err := os.WriteFile(file, []byte("second"), 0o600); err != nil {
		t.Fatal(err)
	}
	if evidence := next.execution.Verify(context.Background()); evidence.Error != "" {
		t.Fatal(evidence.Error)
	}
}

func TestNativeSessionOnlyOneOwnedTurn(t *testing.T) {
	inFlight := map[string]*runningTask{"first": {provider: "claude", nativeSessionID: "session"}, "current": {provider: "claude", nativeSessionID: "session"}}
	if !nativeSessionInUse(inFlight, "current", "claude", "session") {
		t.Fatal("concurrent resume allowed")
	}
	if nativeSessionInUse(inFlight, "current", "codex", "session") || nativeSessionInUse(inFlight, "current", "claude", "") {
		t.Fatal("independent session blocked")
	}
	delete(inFlight, "first")
	if nativeSessionInUse(inFlight, "current", "claude", "session") {
		t.Fatal("own session blocked")
	}
}

func TestNativeSessionConflictingPinIsNotReported(t *testing.T) {
	a, status := historyActor(t.TempDir())
	a.cfg.Reporter = newReporterProbe()
	prepareAndPinHistory(t, a, status, historyRequest("first", ""))
	req := historyRequest("other", "")
	req.Metadata["agent_id"] = "other-agent"
	prepared, err := a.prepareWorkspace(context.Background(), status, req)
	if err != nil {
		t.Fatal(err)
	}
	cancelled := false
	task := &runningTask{taskID: req.ID, request: req, workspace: prepared.workspace, events: prepared.events, cancel: func() { cancelled = true }}
	inFlight := map[string]*runningTask{req.ID: task}
	var detached []detachedReport
	a.handleTaskEvent(context.Background(), &taskEventMsg{taskID: req.ID, event: agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: "session"}}, inFlight, &detached)
	if !cancelled || task.sessionBindingErr == nil {
		t.Fatal("conflicting pin not blocked")
	}
	result := a.recordTerminalResult(context.Background(), task, agentbridge.Result{Status: agentbridge.ResultCompleted, SessionID: "session"})
	if result.Status != agentbridge.ResultBlocked || result.SessionID != "" {
		t.Fatalf("later success bypassed pin failure: %+v", result)
	}
	reporter := a.cfg.Reporter.(*reporterProbe)
	for len(reporter.events) > 0 {
		if event := <-reporter.events; event.Kind == agentbridge.EventSessionIdentified {
			t.Fatal("foreign pin reported")
		}
	}
}

func TestNativeSessionUnsupportedProviderDoesNotClaimHistory(t *testing.T) {
	a, status := historyActor(t.TempDir())
	status.Capabilities[0].SupportsResume = false
	prepareAndPinHistory(t, a, status, historyRequest("first", ""))
	store := a.cfg.Workdir.(workdir.NativeSessions)
	if _, found, err := store.LookupNativeSession("claude", "session"); err != nil || found {
		t.Fatalf("unsupported resume provider claimed history: %v %v", found, err)
	}
	if _, err := a.prepareWorkspace(context.Background(), status, historyRequest("next", "session")); !errors.Is(err, errNativeSessionBinding) {
		t.Fatalf("unsupported resume executed: %v", err)
	}
}

func TestNativeSessionLegacyWarningAndCorruptionBoundary(t *testing.T) {
	root := t.TempDir()
	a, status := historyActor(root)
	if _, err := a.prepareWorkspace(context.Background(), status, historyRequest("legacy", "legacy-session")); err != nil {
		t.Fatal(err)
	}
	select {
	case msg := <-a.mailbox:
		if msg.taskReport == nil || msg.taskReport.event.Kind != agentbridge.EventWarning {
			t.Fatal("missing legacy warning")
		}
	default:
		t.Fatal("legacy continuity silently assumed")
	}
	prepareAndPinHistory(t, a, status, historyRequest("first", ""))
	files, err := os.ReadDir(filepath.Join(root, ".native-sessions"))
	if err != nil || len(files) != 1 {
		t.Fatalf("bindings=%v %v", files, err)
	}
	if err := os.WriteFile(filepath.Join(root, ".native-sessions", files[0].Name()), []byte("broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := a.prepareWorkspace(context.Background(), status, historyRequest("next", "session")); !errors.Is(err, errNativeSessionBinding) {
		t.Fatalf("corrupt history silently replaced: %v", err)
	}
}

func TestNativeSessionResumeRejectsSymlinkedHistory(t *testing.T) {
	a, status := historyActor(t.TempDir())
	first := prepareAndPinHistory(t, a, status, historyRequest("first", ""))
	if err := os.RemoveAll(first.workspace.Workdir); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), first.workspace.Workdir); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := a.prepareWorkspace(context.Background(), status, historyRequest("next", "session")); !errors.Is(err, errNativeSessionBinding) {
		t.Fatalf("escaped history accepted: %v", err)
	}
}

func TestNativeSessionAssignmentCloneIsNotRepeated(t *testing.T) {
	originalClone, originalResolve := runAssignmentGitClone, resolveAssignmentGitExecutable
	t.Cleanup(func() { runAssignmentGitClone = originalClone; resolveAssignmentGitExecutable = originalResolve })
	resolveAssignmentGitExecutable = func(string, string) (string, bool) { return "fake-git", true }
	clones := 0
	runAssignmentGitClone = func(_ context.Context, _ string, args []string) error {
		clones++
		return os.MkdirAll(filepath.Join(args[len(args)-1], ".git"), 0o700)
	}
	a, status := historyActor(t.TempDir())
	firstReq := historyRequest("first", "")
	firstReq.Worktree = publicBranchWorktree()
	first := prepareAndPinHistory(t, a, status, firstReq)
	nextReq := historyRequest("next", "session")
	nextReq.Worktree = publicBranchWorktree()
	next := prepareAndPinHistory(t, a, status, nextReq)
	if clones != 1 || next.workspace.Workdir != first.workspace.Workdir {
		t.Fatalf("clones=%d cwd=%s", clones, next.workspace.Workdir)
	}
	if err := os.RemoveAll(filepath.Join(first.workspace.Workdir, ".git")); err != nil {
		t.Fatal(err)
	}
	lastReq := historyRequest("last", "session")
	lastReq.Worktree = publicBranchWorktree()
	if _, err := a.prepareWorkspace(context.Background(), status, lastReq); !errors.Is(err, errNativeSessionBinding) {
		t.Fatalf("deleted clone silently recreated: %v", err)
	}
	if clones != 1 {
		t.Fatal("recloned missing history")
	}
}
