package main

import (
	"context"
	"os/exec"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process/execwindow"
)

func runtimeModelCommandOutput(executable string, args ...string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	execwindow.Hide(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	return out
}
