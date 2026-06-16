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
// Dependency direction: agentbridge → contracts vocabulary + stdlib. Concrete
// provider adapters depend on agentbridge to implement Adapter. agentbridge
// MUST NOT import any provider/<name> package, and MUST NOT import os/exec,
// net/http, or any filesystem implementation.
package agentbridge

import contractrunstate "github.com/teamswyg/riido-contracts/runstate"

// RunState is one of the 14 run-scope sub-states of a single provider
// session. A task's task-scope StateRunning may map to any non-terminal
// RunState; a run reaching a terminal RunState informs the task FSM
// through bridge events (delivered by the session actor).
//
// Naming is intentionally provider-neutral: docs/20-domain/provider-runtime.md
// forbids provider names in this enum.
type RunState = contractrunstate.RunState

const (
	StatePending             = contractrunstate.StatePending
	StatePreparing           = contractrunstate.StatePreparing
	StateStartingProvider    = contractrunstate.StateStartingProvider
	StateHandshaking         = contractrunstate.StateHandshaking
	StateRunning             = contractrunstate.StateRunning
	StateWaitingToolApproval = contractrunstate.StateWaitingToolApproval
	StateToolRunning         = contractrunstate.StateToolRunning
	StateWaitingProvider     = contractrunstate.StateWaitingProvider
	StateCompleting          = contractrunstate.StateCompleting
	StateCompleted           = contractrunstate.StateCompleted
	StateFailed              = contractrunstate.StateFailed
	StateCancelled           = contractrunstate.StateCancelled
	StateTimedOut            = contractrunstate.StateTimedOut
	StateIdleStopped         = contractrunstate.StateIdleStopped
)

func AllStates() []RunState {
	return contractrunstate.AllStates()
}
