package main

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

func runLocalCommand(timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return string(out), ctx.Err()
	}
	return string(out), err
}

func outputTail(out string) string {
	out = strings.TrimSpace(out)
	if len(out) <= 600 {
		return out
	}
	return out[len(out)-600:]
}
