package supervisor

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

var errAssignmentWorktreeBlocked = errors.New("supervisor: assignment worktree blocked")

func assignmentCloneURL(worktree *assignmentcontract.AssignmentWorktree) (string, error) {
	if worktree == nil {
		return "", nil
	}
	if worktree.IsPrivate {
		return "", fmt.Errorf("%w: private assignment worktree requires git credentials", errAssignmentWorktreeBlocked)
	}
	repoURL, err := assignmentRepositoryURL(worktree)
	if err != nil || repoURL == "" {
		return repoURL, err
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

func assignmentRepositoryURL(worktree *assignmentcontract.AssignmentWorktree) (string, error) {
	repoURL := strings.TrimSpace(worktree.RepositoryURL)
	if repoURL != "" {
		return repoURL, nil
	}
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
	return "https://github.com/" + fullName, nil
}
