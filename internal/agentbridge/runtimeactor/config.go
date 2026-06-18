package runtimeactor

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// Config configures a new Actor.
type Config struct {
	RuntimeID      string
	Owner          string
	DeviceName     string
	Agents         []AgentStatus
	Models         []RuntimeModel
	Adapters       []agentbridge.Adapter
	Process        process.Process
	MaxConcurrent  int
	HeartbeatEvery time.Duration // reserved for the supervisor's heartbeat loop
	HardTimeout    time.Duration // forwarded to each session as its hard timeout
	MailboxSize    int
	// AutoApprove is forwarded to each session.
	AutoApprove agentbridge.AutoApprover
	// ToolStartGate is forwarded to each session.
	ToolStartGate agentbridge.ToolStartGate
	// ToolApprovalGate is forwarded to each session for headless approval
	// requests that were not auto-approved.
	ToolApprovalGate agentbridge.ToolApprovalGate
	// ToolApprovalResolver is forwarded to each session for SaaS/web approval.
	ToolApprovalResolver agentbridge.ToolApprovalResolver
	// PolicyBundleVersion is the active policy bundle version used as a
	// CapabilityFingerprint input. Until C7 grows a policy loader, the
	// daemon supplies a Factor-12 env value and this default keeps local
	// development deterministic.
	PolicyBundleVersion string
	// DetectEnv is passed to each adapter's Detect during Start.
	DetectEnv agentbridge.DetectEnv
	// CapabilityRefreshEvery bounds how long provider detection can stay
	// cached before Submit re-checks the adapter. Set a negative value to
	// disable submit-time refresh.
	CapabilityRefreshEvery time.Duration
	// Now is injected for deterministic tests; defaults to time.Now.
	Now func() time.Time
}
