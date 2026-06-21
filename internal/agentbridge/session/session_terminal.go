package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// TerminateWithContext ends the session with a domain terminal result.
func (s *Session) TerminateWithContext(ctx context.Context, result agentbridge.Result) {
	if ctx == nil {
		ctx = context.Background()
	}
	if result.Status == "" {
		result.Status = agentbridge.ResultFailed
	}
	select {
	case s.terminal <- terminalRequest{ctx: ctx, result: result}:
	default:
	}
}
