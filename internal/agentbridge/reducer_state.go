package agentbridge

import "time"

// State is the reducer's accumulator. It is owned by a single goroutine
// (the session actor); the reducer is a pure function, so there is no
// shared mutation and no mutex.
type State struct {
	Phase                RunState
	SessionID            string
	HasProviderResult    bool
	Usage                Usage
	Tools                map[string]ToolRef
	Output               []byte
	LastSemanticActivity time.Time
	Result               Result
	Terminal             bool
}

// NewState returns an empty State in the Pending phase.
func NewState() State {
	return State{
		Phase: StatePending,
		Tools: map[string]ToolRef{},
	}
}

// AutoApprover decides whether to auto-approve a tool invocation when
// the provider raises EventToolApprovalNeeded. When nil, the reducer
// leaves the run in StateWaitingToolApproval — explicit human approval
// (or an external command from the session actor) is required.
//
// The nil-default enforces the C7 policy stance: unsafe provider permission
// bypass is never implicit.
type AutoApprover func(tool ToolRef) bool
