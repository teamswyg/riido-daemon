package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/claude"
)

type bridgeClaudeAdapter struct{}

func (bridgeClaudeAdapter) Name() string { return claude.Name }

func (bridgeClaudeAdapter) Detect(
	ctx context.Context,
	env agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	return claude.Detect(ctx, env)
}

func (bridgeClaudeAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return claude.BuildStart(req, claude.StartOptions{PermissionMode: claude.PermissionModeApproval})
}

func (bridgeClaudeAdapter) NewParser() agentbridge.Parser { return claude.NewParser() }

func (bridgeClaudeAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return claude.Translate(raw)
}

func (bridgeClaudeAdapter) BlockedArgs() []string { return claude.BlockedArgs() }

func (bridgeClaudeAdapter) BuildProviderInput(cmd agentbridge.Command) ([]byte, error) {
	return claude.BuildProviderInput(cmd)
}

func (bridgeClaudeAdapter) NewProtocolDriver(req agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return claude.NewProtocolDriver(req)
}
