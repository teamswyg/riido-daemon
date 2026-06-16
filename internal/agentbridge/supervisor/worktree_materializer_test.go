package supervisor

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestAssignmentCloneURLRejectsPrivateWorktree(t *testing.T) {
	_, err := assignmentCloneURL(&assignmentcontract.AssignmentWorktree{
		RepositoryFullName: "teamswyg/private-repo",
		RepositoryURL:      "https://github.com/teamswyg/private-repo",
		IsPrivate:          true,
	})
	if err == nil || !strings.Contains(err.Error(), "private assignment worktree requires git credentials") {
		t.Fatalf("expected private repository fail-closed error, got %v", err)
	}
}

func TestAssignmentCloneURLRejectsUnsupportedRepositoryURL(t *testing.T) {
	_, err := assignmentCloneURL(&assignmentcontract.AssignmentWorktree{
		RepositoryURL: "https://token:secret@example.com/teamswyg/riido-daemon",
	})
	if err == nil {
		t.Fatal("expected unsupported URL error")
	}
	if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "token") {
		t.Fatalf("error leaked URL userinfo: %v", err)
	}
}

func TestMaterializeAssignmentWorktreeRunsShallowBranchClone(t *testing.T) {
	var gotGit string
	var gotArgs []string
	original := runAssignmentGitClone
	runAssignmentGitClone = func(_ context.Context, git string, args []string) error {
		gotGit = git
		gotArgs = append([]string(nil), args...)
		return nil
	}
	t.Cleanup(func() { runAssignmentGitClone = original })

	err := materializeAssignmentWorktree(context.Background(), "/tmp/riido-workdir", &assignmentcontract.AssignmentWorktree{
		RepositoryFullName: "teamswyg/riido-daemon",
		RepositoryURL:      "https://github.com/teamswyg/riido-daemon",
		BranchName:         "RIID-4964-agent-profile-upload",
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "git executable not found") {
			t.Skipf("git not available in test environment: %v", err)
		}
		t.Fatalf("materialize worktree: %v", err)
	}
	if gotGit == "" {
		t.Fatal("git executable was not resolved")
	}
	wantArgs := []string{
		"clone",
		"--depth=1",
		"--branch",
		"RIID-4964-agent-profile-upload",
		"https://github.com/teamswyg/riido-daemon",
		"/tmp/riido-workdir",
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("clone args = %#v, want %#v", gotArgs, wantArgs)
	}
}
