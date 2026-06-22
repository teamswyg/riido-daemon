package openclaw

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func buildCommandArgs(req agentbridge.StartRequest, sessionID string) ([]string, []string) {
	args := []string{
		"agent",
		"--local",
		"--json",
		"--session-id", sessionID,
	}
	args = append(args, "--message", buildMessage(req.SystemPrompt, req.Prompt))

	kept, dropped := agentbridge.FilterBlockedArgs(req.CustomArgs, BlockedArgs())
	return append(args, kept...), dropped
}

func buildEnv(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}
