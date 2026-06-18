package validation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
)

func summarize(command string, exitCode int, output []byte, runErr error) string {
	if errors.Is(runErr, context.DeadlineExceeded) {
		return fmt.Sprintf("validation command timed out: %s", command)
	}
	trimmed := string(bytes.TrimSpace(output))
	if len(trimmed) > 400 {
		trimmed = trimmed[:400]
	}
	if trimmed == "" {
		return fmt.Sprintf("validation command exited %d: %s", exitCode, command)
	}
	return fmt.Sprintf("validation command exited %d: %s: %s", exitCode, command, trimmed)
}
