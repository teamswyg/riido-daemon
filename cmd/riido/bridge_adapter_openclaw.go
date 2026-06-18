package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

type bridgeOpenClawAdapter struct{}

func (bridgeOpenClawAdapter) Name() string { return openclaw.Name }

func (bridgeOpenClawAdapter) Detect(
	ctx context.Context,
	env agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	return openclaw.Detect(ctx, env)
}

func (bridgeOpenClawAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return openclaw.BuildStart(req, openclaw.StartOptions{})
}

func (bridgeOpenClawAdapter) NewParser() agentbridge.Parser { return openclaw.NewParser() }

func (bridgeOpenClawAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return openclaw.Translate(raw)
}

func (bridgeOpenClawAdapter) BlockedArgs() []string { return openclaw.BlockedArgs() }
