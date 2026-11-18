package supervisor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestFailedPullCheckoutKeepsIsolatedSessionAcrossRestart(t *testing.T) {
	oldGit, oldClone := resolveAssignmentGitExecutable, runAssignmentGitClone
	t.Cleanup(func() { resolveAssignmentGitExecutable, runAssignmentGitClone = oldGit, oldClone })
	resolveAssignmentGitExecutable = func(string, string) (string, bool) { return "git", true }
	var fetches int
	runAssignmentGitClone = func(_ context.Context, _ string, args []string) error {
		if args[0] == "init" {
			return os.Mkdir(filepath.Join(args[2], ".git"), 0o700)
		}
		if args[2] == "fetch" {
			fetches++
			return classifyAssignmentGitCloneFailure([]byte("could not read Username: fixture-secret"))
		}
		return nil
	}
	newRequest := func(run, session string) *bridge.TaskRequest {
		req := historyRequest(run, session)
		req.Worktree = privateWorktreeRequest().Worktree
		req.Worktree.BranchName = "refs/pull/1719/head"
		req.ProviderNeutralPrompt = "Review PR"
		return req
	}
	root := t.TempDir()
	a, status := historyActor(root)
	first := newRequest("first", "")
	prepared := prepareAndPinHistory(t, a, status, first)
	if !first.AssignmentCheckoutUnavailable || !strings.Contains(first.Prompt, "GitHub tools") || !strings.Contains(first.ProviderNeutralPrompt, "GitHub tools") || strings.Contains(first.Prompt, "fixture-secret") {
		t.Fatal("missing or unsafe checkout guidance")
	}
	if _, err := os.Stat(filepath.Join(first.Cwd, ".git")); !os.IsNotExist(err) {
		t.Fatalf("partial PR checkout remained: %v", err)
	}
	if stages, err := filepath.Glob(filepath.Join(prepared.workspace.Root, ".checkout-*")); err != nil || len(stages) != 0 {
		t.Fatal("staging directory remained", stages, err)
	}
	a, status = historyActor(root)
	next := newRequest("second", "session")
	prepareAndPinHistory(t, a, status, next)
	if !next.AssignmentCheckoutUnavailable || next.Cwd != first.Cwd || fetches != 1 {
		t.Fatal("lost the isolated session on resume")
	}
	next.Worktree.BranchName = "other"
	if _, err := a.prepareWorkspace(t.Context(), status, next); !errors.Is(err, errNativeSessionBinding) {
		t.Fatalf("changed repository grant resumed: %v", err)
	}
}
