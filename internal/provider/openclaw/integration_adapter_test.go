package openclaw

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type integrationAdapter struct {
	sessionID string
}

func (integrationAdapter) Name() string { return Name }

func (integrationAdapter) Detect(
	_ context.Context,
	_ agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}

func (a integrationAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return BuildStart(req, StartOptions{SessionID: a.sessionID})
}

func (integrationAdapter) NewParser() agentbridge.Parser { return NewParser() }

func (integrationAdapter) Translate(
	raw agentbridge.RawEvent,
) ([]agentbridge.Event, []agentbridge.Command, error) {
	return Translate(raw)
}

func (integrationAdapter) BlockedArgs() []string { return BlockedArgs() }
