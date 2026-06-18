package claude

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// integrationAdapter wraps the package-level helpers as an Adapter so
// session.Start can use it. We can't reach into the bridge package
// here without an import cycle, so we keep this duplicate small.
type integrationAdapter struct{}

func (integrationAdapter) Name() string { return Name }

func (integrationAdapter) Detect(
	_ context.Context,
	_ agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}

func (integrationAdapter) BuildStart(
	req agentbridge.StartRequest,
) (agentbridge.StartCommand, error) {
	return BuildStart(req, StartOptions{PermissionMode: PermissionModeApproval})
}

func (integrationAdapter) NewParser() agentbridge.Parser { return NewParser() }

func (integrationAdapter) Translate(
	raw agentbridge.RawEvent,
) ([]agentbridge.Event, []agentbridge.Command, error) {
	return Translate(raw)
}

func (integrationAdapter) BlockedArgs() []string { return BlockedArgs() }
