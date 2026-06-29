package main

import (
	"context"
	"os/exec"
	"time"
)

func commandOutput(executable string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, executable, args...).Output()
}
