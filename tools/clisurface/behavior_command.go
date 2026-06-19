package main

import (
	"context"
	"os/exec"
	"time"
)

func runGoCommand(repo string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	return string(out), err
}
