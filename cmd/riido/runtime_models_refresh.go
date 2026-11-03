package main

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process/execwindow"
)

// The embedded catalog keeps startup responsive. Live discovery replaces it
// from the actor's background probe, preserving the prior catalog on failure.
func refreshDaemonRuntimeModels(ctx context.Context, capability runtimeactor.Capability) []runtimeactor.RuntimeModel {
	var args []string
	switch capability.Provider {
	case "cursor":
		args = []string{"models"}
	case "claude":
		args = []string{"--help"}
	default:
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, capability.Executable, args...)
	execwindow.Hide(cmd)
	body, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}
	switch capability.Provider {
	case "cursor":
		return parseCursorRuntimeModelList(body, cursorRuntimeDefaultModel(os.UserHomeDir).ModelID)
	case "claude":
		return parseClaudeRuntimeModelHelp(body)
	}
	return nil
}
