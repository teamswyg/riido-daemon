package session

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

const (
	// DefaultEventBuffer is the C4 Provider Runtime backpressure buffer
	// recorded in docs/20-domain/provider-runtime.md §7.5.
	DefaultEventBuffer = 256
	// DefaultResultBuffer stores the single terminal result for a session.
	DefaultResultBuffer = 1
	// DefaultProcessKillTimeout bounds provider process kill calls made during
	// session teardown and reducer-driven cancellation.
	DefaultProcessKillTimeout = 5 * time.Second
)

// Config is the input to Start.
type Config struct {
	TaskID    string
	RuntimeID string

	Adapter agentbridge.Adapter
	Process process.Process
	Spawn   process.Command
	Request agentbridge.StartRequest

	// HardTimeout is the wall-clock deadline for the entire run. Zero disables.
	HardTimeout time.Duration
	// SemanticIdle is the idle watchdog timeout. Zero disables.
	SemanticIdle time.Duration

	// AutoApprove decides tool approval. Default nil requires human approval.
	AutoApprove agentbridge.AutoApprover
	// ToolStartGate fails closed if a started tool cannot be approved.
	ToolStartGate agentbridge.ToolStartGate
	// ToolApprovalGate fails closed for headless approval requests.
	ToolApprovalGate agentbridge.ToolApprovalGate
	// ToolApprovalResolver connects approval requests to a human decision loop.
	ToolApprovalResolver agentbridge.ToolApprovalResolver

	// EventBuffer / ResultBuffer override default channel capacities.
	EventBuffer  int
	ResultBuffer int

	// ProtocolDriver optionally owns raw-event routing for active handshakes.
	ProtocolDriver ProtocolDriver

	// TempFiles are adapter-owned files removed after exit/cancel/timeout.
	TempFiles []string

	// ProcessKillTimeout bounds RunningProcess.Kill calls.
	ProcessKillTimeout time.Duration

	// now is injected for deterministic tests; defaults to time.Now.
	Now func() time.Time
}
