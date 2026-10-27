package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
)

// A missing source scope permits legacy work but cannot produce positive evidence.
func beginClaimHostWork(ctx context.Context) (*executionproof.HostWork, error) {
	scope := executionproof.ScopeFromContext(ctx)
	if scope == nil {
		return nil, nil
	}
	return scope.Tracker().BeginHostWork()
}
