package supervisor

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

var (
	resolveAssignmentGitExecutable = detectutil.ResolveExecutable
	runAssignmentGitClone          = defaultRunAssignmentGitClone
)

func materializeAssignmentWorktree(ctx context.Context, targetDir string, worktree *assignmentcontract.AssignmentWorktree, onPullCheckout func(string)) error {
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
		return errors.New("supervisor: git executable not found for assignment worktree. Action: Install Git on the daemon device, confirm `git --version` works for the same OS account as Riido, then restart Riido and retry the task")
	}
	branch := strings.TrimSpace(worktree.BranchName)
	fallback := assignmentPullRequestFallback(worktree.Source)
	var cloneErr error
	var pulledRef string
	if assignmentPullRequestHead.MatchString(branch) || (branch == "" && fallback != "") {
		ref := branch
		if ref == "" {
			ref = fallback
		}
		pulledRef = ref
		cloneErr = materializeAssignmentPullRequest(ctx, git, cloneURL, targetDir, ref)
	} else {
		args := []string{"clone", "--depth=1"}
		if branch != "" {
			args = append(args, "--branch", branch)
		}
		cloneErr = runAssignmentGitClone(ctx, git, append(args, cloneURL, targetDir))
		var diagnostic *assignmentGitCloneFailure
		if fallback != "" && ctx.Err() == nil && errors.As(cloneErr, &diagnostic) && diagnostic.branchMissing {
			pulledRef = fallback
			cloneErr = materializeAssignmentPullRequest(ctx, git, cloneURL, targetDir, fallback)
		}
	}
	if err := cloneErr; err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if worktree.IsPrivate {
			var diagnostic *assignmentGitCloneFailure
			if !errors.As(err, &diagnostic) {
				diagnostic = classifyAssignmentGitCloneFailure(nil)
			}
			return fmt.Errorf("%w: %w", errAssignmentWorktreeBlocked, diagnostic)
		}
		return fmt.Errorf("supervisor: clone assignment worktree: %w", err)
	}
	if pulledRef != "" && onPullCheckout != nil {
		onPullCheckout(strings.TrimSuffix(strings.TrimPrefix(pulledRef, "refs/pull/"), "/head"))
	}
	return nil
}

// A GitHub pull ref can outlive a deleted head branch. Fetch exactly the
// selected ref into a local worktree; never substitute the default branch.
var (
	assignmentPullRequestHead           = regexp.MustCompile(`^refs/pull/[1-9][0-9]{0,9}/head$`)
	assignmentPullRequestFallbackSource = regexp.MustCompile(`^pull_request_fallback:([1-9][0-9]{0,9})$`)
)

func assignmentPullRequestFallback(source string) string {
	match := assignmentPullRequestFallbackSource.FindStringSubmatch(source)
	if match == nil {
		return ""
	}
	return "refs/pull/" + match[1] + "/head"
}

func materializeAssignmentPullRequest(ctx context.Context, git, cloneURL, dir, ref string) error {
	for _, args := range [][]string{
		{"init", "--", dir},
		{"-C", dir, "remote", "add", "origin", cloneURL},
		{"-C", dir, "fetch", "--depth=1", "--no-tags", "origin", ref},
		{"-C", dir, "checkout", "--detach", "FETCH_HEAD"},
	} {
		if err := runAssignmentGitClone(ctx, git, args); err != nil {
			return err
		}
	}
	return nil
}
