package supervisor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestAssignmentGitCloneDoesNotExposeHelperOutput(t *testing.T) {
	if os.Getenv("RIIDO_TEST_GIT_OUTPUT_HELPER") == "1" {
		fmt.Fprintln(os.Stdout, "fixture-credential-stdout")
		fmt.Fprintln(os.Stderr, "fixture-credential-stderr")
		os.Exit(1)
	}
	t.Setenv("RIIDO_TEST_GIT_OUTPUT_HELPER", "1")
	err := defaultRunAssignmentGitClone(context.Background(), os.Args[0], []string{"-test.run=^TestAssignmentGitCloneDoesNotExposeHelperOutput$"})
	if err == nil || strings.Contains(err.Error(), "fixture-credential") {
		t.Fatalf("Git failure must not include helper output: %v", err)
	}
}

func TestPrivateWorktreeCancellationRemainsCancellation(t *testing.T) {
	original := runAssignmentGitClone
	runAssignmentGitClone = func(ctx context.Context, _ string, _ []string) error { return ctx.Err() }
	t.Cleanup(func() { runAssignmentGitClone = original })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := materializeAssignmentWorktree(ctx, t.TempDir(), privateWorktreeRequest().Worktree)
	if !errors.Is(err, context.Canceled) || errors.Is(err, errAssignmentWorktreeBlocked) {
		t.Fatalf("canceled private preparation = %v", err)
	}
}
