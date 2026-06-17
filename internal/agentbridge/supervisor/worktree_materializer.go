package supervisor

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

var (
	resolveAssignmentGitExecutable = detectutil.ResolveExecutable
	runAssignmentGitClone          = defaultRunAssignmentGitClone
)

var errAssignmentWorktreeBlocked = errors.New("supervisor: assignment worktree blocked")

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

func assignmentCloneURL(worktree *assignmentcontract.AssignmentWorktree) (string, error) {
	if worktree == nil {
		return "", nil
	}
	if worktree.IsPrivate {
		return "", fmt.Errorf("%w: private assignment worktree requires git credentials", errAssignmentWorktreeBlocked)
	}
	repoURL := strings.TrimSpace(worktree.RepositoryURL)
	if repoURL == "" {
		rawFullName := strings.TrimSpace(worktree.RepositoryFullName)
		if rawFullName == "" {
			return "", nil
		}
		fullName := assignmentcontract.NormalizePublicGitHubRepositoryFullName(rawFullName)
		if fullName == "" {
			parsed := &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "/" + strings.Trim(rawFullName, "/"),
			}
			return "", fmt.Errorf("%w: unsupported assignment repository_url %q", errAssignmentWorktreeBlocked, redactedRepositoryURL(parsed))
		}
		repoURL = "https://github.com/" + fullName
	}
	parsed, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("%w: invalid assignment repository_url: %w", errAssignmentWorktreeBlocked, err)
	}
	cloneURL := assignmentcontract.NormalizePublicGitHubRepositoryURL(repoURL)
	if cloneURL == "" {
		return "", fmt.Errorf("%w: unsupported assignment repository_url %q", errAssignmentWorktreeBlocked, redactedRepositoryURL(parsed))
	}
	return cloneURL, nil
}

func redactedRepositoryURL(parsed *url.URL) string {
	if parsed == nil {
		return ""
	}
	copyURL := *parsed
	copyURL.User = nil
	copyURL.RawQuery = ""
	copyURL.ForceQuery = false
	copyURL.Fragment = ""
	copyURL.RawFragment = ""
	copyURL.RawPath = ""
	if fullName := assignmentcontract.NormalizePublicGitHubRepositoryFullName(copyURL.Path); fullName != "" {
		copyURL.Path = "/" + fullName
	} else {
		copyURL.Path = "/redacted"
	}
	return copyURL.String()
}

func defaultRunAssignmentGitClone(ctx context.Context, git string, args []string) error {
	cmd := exec.CommandContext(ctx, git, args...)
	cmd.Env = append(detectutil.EnvListWithLaunchPATH(os.Environ(), ""), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
