// Package agentbridge is the ACL between provider CLI processes and
// Riido's canonical event log. It owns the C4 Provider Runtime/Adapter
// port (Adapter interface, StartCommand, raw event envelope) and the
// run-scope sub-FSM that lives inside a task's Running state.
//
// What this package owns:
//   - Adapter port + RawEvent / Parser / DetectResult
//   - RunState (14 provider-neutral run-scope sub-states)
//   - EventKind (provider-neutral run-scope event catalog)
//   - CommandKind (reducer-emitted imperatives)
//   - The reducer + the State accumulator
//   - FilterBlockedArgs helper for protocol-critical arg enforcement
//
// What this package does NOT own:
//   - Task-scope FSM (Created/Queued/Validating/...) → public
//     riido-contracts/task (C1).
//   - The canonical event log envelope → public riido-contracts/ir (C2).
//   - Provider capability model (ProviderCapability/Fingerprint) → public
//     riido-contracts/provider/capability (C3).
//   - Concrete adapter implementations → future provider migration slices.
//   - Process spawning → internal/process.
//
// Dependency direction: agentbridge → (stdlib only). Concrete provider
// adapters depend on agentbridge to implement Adapter. agentbridge MUST
// NOT import any provider/<name> package, and MUST NOT import os/exec,
// net/http, or any filesystem implementation.
package agentbridge

// RunState is one of the 14 run-scope sub-states of a single provider
// session. A task's task-scope StateRunning may map to any non-terminal
// RunState; a run reaching a terminal RunState informs the task FSM
// through bridge events (delivered by the session actor).
//
// Naming is intentionally provider-neutral: docs/20-domain/provider-runtime.md
// forbids provider names in this enum.
type RunState string

const (
	StatePending             RunState = "pending"
	StatePreparing           RunState = "preparing"
	StateStartingProvider    RunState = "starting_provider"
	StateHandshaking         RunState = "handshaking"
	StateRunning             RunState = "running"
	StateWaitingToolApproval RunState = "waiting_tool_approval"
	StateToolRunning         RunState = "tool_running"
	StateWaitingProvider     RunState = "waiting_provider"
	StateCompleting          RunState = "completing"
	StateCompleted           RunState = "completed"
	StateFailed              RunState = "failed"
	StateCancelled           RunState = "cancelled"
	StateTimedOut            RunState = "timed_out"
	StateIdleStopped         RunState = "idle_stopped"
)

func AllStates() []RunState {
	return []RunState{
		StatePending, StatePreparing, StateStartingProvider,
		StateHandshaking, StateRunning, StateWaitingToolApproval,
		StateToolRunning, StateWaitingProvider, StateCompleting,
		StateCompleted, StateFailed, StateCancelled, StateTimedOut,
		StateIdleStopped,
	}
}

// IsTerminal reports whether s is one of the five terminal run-scope
// states. No transition can originate from a terminal RunState
// (reducer invariant 1).
func (s RunState) IsTerminal() bool {
	switch s {
	case StateCompleted, StateFailed, StateCancelled, StateTimedOut, StateIdleStopped:
		return true
	default:
		return false
	}
}
