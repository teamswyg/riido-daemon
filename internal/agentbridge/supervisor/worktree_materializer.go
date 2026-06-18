package supervisor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

var (
	resolveAssignmentGitExecutable = detectutil.ResolveExecutable
	runAssignmentGitClone          = defaultRunAssignmentGitClone
)

func materializeAssignmentWorktree(ctx context.Context, targetDir string, worktree *assignmentcontract.AssignmentWorktree) error {
	if worktree == nil {
		return nil
	}
	cloneURL, err := assignmentCloneURL(worktree)
	if err != nil {
		return err
	}
	if cloneURL == "" {
		return nil
	}
	git, ok := resolveAssignmentGitExecutable("git", "")
	if !ok {
		return errors.New("supervisor: git executable not found for assignment worktree")
	}
	args := []string{"clone", "--depth=1"}
	if branch := strings.TrimSpace(worktree.BranchName); branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, cloneURL, targetDir)
	if err := runAssignmentGitClone(ctx, git, args); err != nil {
		return fmt.Errorf("supervisor: clone assignment worktree: %w", err)
	}
	return nil
}
