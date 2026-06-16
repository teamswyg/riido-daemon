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
		return "", errors.New("supervisor: private assignment worktree requires git credentials")
	}
	repoURL := strings.TrimSpace(worktree.RepositoryURL)
	if repoURL == "" {
		fullName := strings.Trim(strings.TrimSpace(worktree.RepositoryFullName), "/")
		if fullName == "" {
			return "", nil
		}
		repoURL = "https://github.com/" + fullName
	}
	parsed, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("supervisor: invalid assignment repository_url: %w", err)
	}
	if parsed.Scheme != "https" || !strings.EqualFold(parsed.Hostname(), "github.com") || parsed.User != nil {
		return "", fmt.Errorf("supervisor: unsupported assignment repository_url %q", redactedRepositoryURL(parsed))
	}
	if parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || parsed.RawFragment != "" {
		return "", fmt.Errorf("supervisor: unsupported assignment repository_url %q", redactedRepositoryURL(parsed))
	}
	if !validGitHubRepositoryPath(parsed.Path) {
		return "", fmt.Errorf("supervisor: unsupported assignment repository_url %q", redactedRepositoryURL(parsed))
	}
	return parsed.String(), nil
}

func validGitHubRepositoryPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		return false
	}
	for _, part := range parts {
		if !validGitHubRepositoryPart(part) {
			return false
		}
	}
	return true
}

func validGitHubRepositoryPart(part string) bool {
	if part == "" || part == "." || part == ".." {
		return false
	}
	for _, ch := range part {
		switch {
		case ch >= 'a' && ch <= 'z':
		case ch >= 'A' && ch <= 'Z':
		case ch >= '0' && ch <= '9':
		case ch == '-' || ch == '_' || ch == '.':
		default:
			return false
		}
	}
	return true
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
	if !validGitHubRepositoryPath(copyURL.Path) {
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
