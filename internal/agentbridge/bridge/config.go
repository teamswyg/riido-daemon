package bridge

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// Config carries dependencies that the caller supplies.
type Config struct {
	// Adapters MUST contain at least one provider adapter.
	Adapters []agentbridge.Adapter
	// Process is the spawn port. When nil the bridge has no way to spawn;
	// production callers supply a real os/exec adapter.
	Process process.Process
	// DefaultTimeout applies when TaskRequest.Timeout is zero.
	DefaultTimeout time.Duration
	// DefaultSemanticIdle applies when TaskRequest.SemanticIdle is zero.
	DefaultSemanticIdle time.Duration
	// AutoApprove is the session-level tool-approval policy. Nil means human.
	AutoApprove agentbridge.AutoApprover
	// ToolStartGate is the session-level fail-closed policy for tool calls.
	ToolStartGate agentbridge.ToolStartGate
	// ToolApprovalGate is the headless fail-closed policy for approval requests.
	ToolApprovalGate agentbridge.ToolApprovalGate
	// ToolApprovalResolver is the optional external human decision loop.
	ToolApprovalResolver agentbridge.ToolApprovalResolver
}
