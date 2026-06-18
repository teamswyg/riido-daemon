package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
)

type bridgeCodexAdapter struct{}

func (bridgeCodexAdapter) Name() string { return codex.Name }

func (bridgeCodexAdapter) Detect(
	ctx context.Context,
	env agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	return codex.Detect(ctx, env)
}

func (bridgeCodexAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return codex.BuildStart(req, codex.StartOptions{})
}

func (bridgeCodexAdapter) NewParser() agentbridge.Parser { return codex.NewParser() }

func (bridgeCodexAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return codex.Translate(raw)
}

func (bridgeCodexAdapter) BlockedArgs() []string { return codex.BlockedArgs() }

func (bridgeCodexAdapter) NewProtocolDriver(req agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return codex.NewProtocolDriver(req)
}
