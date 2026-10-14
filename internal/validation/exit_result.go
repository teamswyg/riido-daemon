package validation

import (
	"context"
	"errors"
	"os/exec"
)

func exitCodeFor(ctx context.Context, err error) int {
	if ctx.Err() == context.DeadlineExceeded {
		return 124
	}
	if err == nil {
		return 0
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}
	return 1
}

func resultForExitCode(exitCode int) string {
	if exitCode == 0 {
		return "passed"
	}
	return "failed"
}
