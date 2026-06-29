package main

import (
	"context"
	"os/exec"
	"time"
)

func runtimeModelCommandOutput(executable string, args ...string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, executable, args...).Output()
	if err != nil {
		return nil
	}
	return out
}
