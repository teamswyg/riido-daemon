package supervisor

import (
	"context"
	"testing"
)

func stubCancellableAssignmentClone(t *testing.T) (chan struct{}, chan struct{}) {
	t.Helper()
	cloneStarted := make(chan struct{})
	cloneCanceled := make(chan struct{})
	originalGitResolver := resolveAssignmentGitExecutable
	originalClone := runAssignmentGitClone
	resolveAssignmentGitExecutable = func(string, string) (string, bool) {
		return "/usr/bin/git", true
	}
	runAssignmentGitClone = func(ctx context.Context, _ string, _ []string) error {
		close(cloneStarted)
		<-ctx.Done()
		close(cloneCanceled)
		return ctx.Err()
	}
	t.Cleanup(func() {
		runAssignmentGitClone = originalClone
		resolveAssignmentGitExecutable = originalGitResolver
	})
	return cloneStarted, cloneCanceled
}
