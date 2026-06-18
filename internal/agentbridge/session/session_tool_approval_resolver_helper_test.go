package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type resolverFunc func(context.Context, string, agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error)

func (f resolverFunc) ResolveToolApproval(ctx context.Context, executionID string, tool agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
	return f(ctx, executionID, tool)
}
