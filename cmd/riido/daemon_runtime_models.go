package main

import (
	"os"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/provider/claude"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

func daemonRuntimeModels(provider string) []runtimeactor.RuntimeModel {
	switch strings.TrimSpace(provider) {
	case codex.Name:
		return codexRuntimeModels(os.UserHomeDir)
	case cursor.Name:
		return cursorRuntimeModels(os.UserHomeDir)
	case openclaw.Name:
		return openClawRuntimeModels(os.UserHomeDir)
	case claude.Name:
		return nil
	default:
		return nil
	}
}
