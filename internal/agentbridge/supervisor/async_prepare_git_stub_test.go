package supervisor

import (
	"context"
	"sync"
	"testing"
)

func stubSlowAssignmentClone(t *testing.T) (<-chan struct{}, <-chan struct{}) {
	t.Helper()

	cloneStarted := make(chan struct{})
	cloneDone := make(chan struct{})
	unblockClone := make(chan struct{})
	var unblockOnce sync.Once

	originalGitResolver := resolveAssignmentGitExecutable
	originalClone := runAssignmentGitClone
	resolveAssignmentGitExecutable = func(string, string) (string, bool) {
		return "/usr/bin/git", true
	}
	runAssignmentGitClone = func(ctx context.Context, _ string, _ []string) error {
		close(cloneStarted)
		defer close(cloneDone)
		select {
		case <-unblockClone:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	t.Cleanup(func() {
		unblockOnce.Do(func() { close(unblockClone) })
		runAssignmentGitClone = originalClone
		resolveAssignmentGitExecutable = originalGitResolver
	})

	return cloneStarted, cloneDone
}
