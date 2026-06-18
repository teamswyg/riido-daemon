package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
)

type bridgeCursorAdapter struct{}

func (bridgeCursorAdapter) Name() string { return cursor.Name }

func (bridgeCursorAdapter) Detect(
	ctx context.Context,
	env agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	return cursor.Detect(ctx, env)
}

func (bridgeCursorAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return cursor.BuildStart(req, cursor.StartOptions{})
}

func (bridgeCursorAdapter) NewParser() agentbridge.Parser { return cursor.NewParser() }

func (bridgeCursorAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return cursor.Translate(raw)
}

func (bridgeCursorAdapter) BlockedArgs() []string { return cursor.BlockedArgs() }
