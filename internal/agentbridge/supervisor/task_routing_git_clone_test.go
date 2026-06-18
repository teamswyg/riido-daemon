package supervisor

import (
	"context"
	"testing"
)

func captureRunAssignmentGitClone(t *testing.T) *[]string {
	t.Helper()
	var gotCloneArgs []string
	originalClone := runAssignmentGitClone
	runAssignmentGitClone = func(_ context.Context, _ string, args []string) error {
		gotCloneArgs = append([]string(nil), args...)
		return nil
	}
	t.Cleanup(func() { runAssignmentGitClone = originalClone })
	return &gotCloneArgs
}
