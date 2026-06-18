package supervisor

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

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

	err := materializeAssignmentWorktree(
		context.Background(),
		"/tmp/riido-workdir",
		publicBranchWorktree(),
	)
	if err != nil {
		if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "git executable not found") {
			t.Skipf("git not available in test environment: %v", err)
		}
		t.Fatalf("materialize worktree: %v", err)
	}
	if gotGit == "" {
		t.Fatal("git executable was not resolved")
	}
	if !reflect.DeepEqual(gotArgs, shallowBranchCloneArgs()) {
		t.Fatalf("clone args = %#v, want %#v", gotArgs, shallowBranchCloneArgs())
	}
}

func publicBranchWorktree() *assignmentcontract.AssignmentWorktree {
	return &assignmentcontract.AssignmentWorktree{
		RepositoryFullName: "teamswyg/riido-daemon",
		RepositoryURL:      "https://github.com/teamswyg/riido-daemon",
		BranchName:         "RIID-4964-agent-profile-upload",
	}
}

func shallowBranchCloneArgs() []string {
	return []string{
		"clone",
		"--depth=1",
		"--branch",
		"RIID-4964-agent-profile-upload",
		"https://github.com/teamswyg/riido-daemon",
		"/tmp/riido-workdir",
	}
}
