// Package runtimeactor owns the C4/C5 provider-neutral Runtime tier of
// Riido's Daemon -> Runtime -> Agent hierarchy.
//
// One Actor per local runtime capability boundary. The production daemon
// creates one RuntimeActor per provider adapter and the SupervisorActor
// dispatches across that pool. It holds:
//   - A capability snapshot for the registered provider Adapter(s).
//   - A bounded slot pool for this runtime (MaxConcurrent).
//   - The set of currently in-flight SessionActors.
//
// Actor state is owned by a single goroutine. Callers interact through
// bounded mailbox channels (Submit / Cancel / Status / Stop). No mutex
// is used in domain code. This package does not own supervisor task
// claim loops, control-plane transport, task persistence, or concrete
// provider adapters. See docs/20-domain/provider-runtime.md §7.7.
//
// The package is intentionally NOT named `runtime` to avoid colliding
// with Go's stdlib `runtime` package.
package runtimeactor

import (
	"errors"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
)

const (
	// DefaultMailboxSize is the runtime actor mailbox size fixed by
	// docs/20-domain/provider-runtime.md §7.5.
	DefaultMailboxSize = 16
)

// ----- Errors -----

var (
	// ErrUnknownProvider — Submit referenced a provider name the Actor
	// does not have an Adapter for.
	ErrUnknownProvider = errors.New("runtimeactor: unknown provider")
	// ErrSlotExhausted — MaxConcurrent reached. Policy: reject (the
	// caller may retry or queue externally). audit M-1 §HonorsMaxConcurrentSlots.
	ErrSlotExhausted = errors.New("runtimeactor: max concurrent reached")
	// ErrUnknownTask — Cancel referenced a taskID that is not in-flight.
	ErrUnknownTask = errors.New("runtimeactor: unknown task")
	// ErrActorStopped — Submit or Cancel after Stop.
	ErrActorStopped = errors.New("runtimeactor: stopped")
	// ErrProviderUnavailable — the provider's Detect reported Available=false.
	ErrProviderUnavailable = errors.New("runtimeactor: provider unavailable")
	// ErrDuplicateTaskID — Submit with a taskID that is already running.
	ErrDuplicateTaskID = errors.New("runtimeactor: duplicate task id")
)

// ----- Public surface -----

// Capability is the daemon-side view of a single provider's runtime
// readiness. Built from Adapter.Detect.
type Capability struct {
	Provider                  string `json:"provider"`
	Available                 bool   `json:"available"`
	Version                   string `json:"version,omitempty"`
	Executable                string `json:"executable,omitempty"`
	Profile                   string `json:"profile,omitempty"`
	Reason                    string `json:"reason,omitempty"`
	ProtocolKind              string `json:"protocol_kind,omitempty"`
	AdapterID                 string `json:"adapter_id,omitempty"`
	AdapterVersion            string `json:"adapter_version,omitempty"`
	ProtocolVersion           string `json:"protocol_version,omitempty"`
	CompatibilityStatus       string `json:"compatibility_status,omitempty"`
	CapabilityFingerprint     string `json:"capability_fingerprint,omitempty"`
	DetectedFingerprint       string `json:"detected_fingerprint,omitempty"`
	RequiresExperimentalOptIn bool   `json:"requires_experimental_opt_in,omitempty"`
	SupportsStreaming         bool   `json:"supports_streaming"`
	SupportsResume            bool   `json:"supports_resume"`
	SupportsSystem            bool   `json:"supports_system"`
	SupportsMaxTurns          bool   `json:"supports_max_turns"`
	SupportsMCP               bool   `json:"supports_mcp"`
	SupportsToolHooks         bool   `json:"supports_tool_hooks"`
	SupportsUsage             bool   `json:"supports_usage"`
	SupportsFileEvents        bool   `json:"supports_file_events"`
	SupportsWorktree          bool   `json:"supports_worktree"`
}

// TaskStatus describes one in-flight task within the runtime.
type TaskStatus struct {
	TaskID    string `json:"task_id"`
	Provider  string `json:"provider"`
	SessionID string `json:"session_id,omitempty"`
	State     string `json:"state"`
}

// AgentStatus is the runtime -> agent association shown by the local
// settings UI. The runtime actor does not schedule per-agent work yet;
// it simply publishes the binding data supplied by the daemon layer.
type AgentStatus struct {
	AgentID string `json:"agent_id,omitempty"`
	Name    string `json:"name"`
	State   string `json:"state,omitempty"`
}

// RuntimeModel is the runtime-scoped model catalog projected to the
// control plane. Model IDs are opaque to the daemon except for the local
// provider config source that reported them.
type RuntimeModel struct {
	ModelID   string `json:"model_id"`
	Label     string `json:"label"`
	IsDefault bool   `json:"is_default"`
}

// Status is the synchronous Status(ctx) snapshot.
type Status struct {
	RuntimeID       string         `json:"runtime_id"`
	StartedAt       time.Time      `json:"started_at"`
	UptimeSeconds   int64          `json:"uptime_seconds"`
	Health          string         `json:"health"`
	Owner           string         `json:"owner,omitempty"`
	DeviceName      string         `json:"device_name,omitempty"`
	Agents          []AgentStatus  `json:"agents,omitempty"`
	Models          []RuntimeModel `json:"models,omitempty"`
	Capabilities    []Capability   `json:"capabilities"`
	MaxConcurrent   int            `json:"max_concurrent"`
	RunningSessions int            `json:"running_sessions"`
	RunningTasks    []TaskStatus   `json:"running_tasks"`
}

// Heartbeat is the publish-ready payload for ControlPlane.Heartbeat.
type Heartbeat struct {
	RuntimeID      string   `json:"runtime_id"`
	UptimeSeconds  int64    `json:"uptime_seconds"`
	DeviceName     string   `json:"device_name,omitempty"`
	SlotLimit      int      `json:"slot_limit"`
	SlotsInUse     int      `json:"slots_in_use"`
	RunningTaskIDs []string `json:"running_task_ids"`
}

// SessionHandle is the caller-facing per-task handle. Mirrors
// session.Session but is the Actor's surface so we don't leak the
// internal session package across the API boundary.
type SessionHandle struct {
	TaskID  string
	session *session.Session
}

// Events returns the run-scope event stream, closed when the session
// terminates.
func (h *SessionHandle) Events() <-chan agentbridge.Event { return h.session.Events() }

// Result returns the terminal result channel (single value, then closed).
func (h *SessionHandle) Result() <-chan agentbridge.Result { return h.session.Result() }

// Done signals termination without consuming Result. Used by the Actor
// itself; callers normally prefer Result().
func (h *SessionHandle) Done() <-chan struct{} { return h.session.Done() }

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
	// PolicyBundleVersion is the active policy bundle version used as a
	// CapabilityFingerprint input. Until C7 grows a policy loader, the
	// daemon supplies a Factor-12 env value and this default keeps local
	// development deterministic.
	PolicyBundleVersion string
	// DetectEnv is passed to each adapter's Detect during Start.
	DetectEnv agentbridge.DetectEnv
	// Now is injected for deterministic tests; defaults to time.Now.
	Now func() time.Time
}
