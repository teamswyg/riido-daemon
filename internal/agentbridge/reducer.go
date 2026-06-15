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

// Reduce applies a single Event to s and returns the new State plus any
// commands the session actor should execute.
//
// Invariants enforced by docs/20-domain/provider-runtime.md:
//
//  1. Terminal state is entered exactly once.
//  2. After terminal, non-terminal events do not change Phase. Usage
//     may still accumulate.
//  3. SessionIdentified may arrive at any time before terminal.
//  4. Completed is valid with empty Output.
//  5. Process exit code 0 paired with a provider failed result → failed.
//     Process non-zero with no provider result → failed.
//  6. Cancellation preempts any later result or process exit.
//  7. Timeout preempts any later result or process exit.
//  8. RunState carries no provider name.
//  9. Usage accumulates via Usage.Add regardless of provider schema.
//  10. AutoApprover==nil means human approval is required (no implicit
//     bypass).
func Reduce(s State, e Event, approver AutoApprover) (State, []Command) {
	if s.Terminal {
		// Invariant 2: usage may still accumulate; phase stays put.
		if e.Kind == EventUsageDelta {
			s.Usage = s.Usage.Add(e.Usage)
		}
		return s, nil
	}

	if e.Kind.IsSemanticActivity() && !e.At.IsZero() {
		s.LastSemanticActivity = e.At
	}

	switch e.Kind {
	case EventSessionIdentified:
		s.SessionID = e.SessionID
		return s, []Command{{Kind: CommandPersistSession}}

	case EventLifecycle:
		if e.Phase != "" && !e.Phase.IsTerminal() {
			s.Phase = e.Phase
		}
		return s, nil

	case EventToolApprovalNeeded:
		s.Phase = StateWaitingToolApproval
		if e.Tool.ID != "" {
			s.Tools[e.Tool.ID] = e.Tool
		}
		if approver != nil && approver(e.Tool) {
			return s, []Command{{
				Kind:              CommandApproveTool,
				ToolID:            e.Tool.ID,
				ProviderRequestID: e.Tool.ProviderRequestID,
			}}
		}
		return s, nil

	case EventToolCallStarted:
		s.Phase = StateToolRunning
		if e.Tool.ID != "" {
			s.Tools[e.Tool.ID] = e.Tool
		}
		return s, nil

	case EventToolCallDelta:
		return s, nil

	case EventToolCallCompleted, EventToolCallFailed:
		if e.Tool.ID != "" {
			delete(s.Tools, e.Tool.ID)
		}
		if len(s.Tools) == 0 {
			s.Phase = StateRunning
		}
		return s, nil

	case EventUsageDelta:
		s.Usage = s.Usage.Add(e.Usage)
		return s, nil

	case EventTextDelta:
		s.Output = append(s.Output, e.Text...)
		return s, nil

	case EventThinkingDelta, EventProgress, EventLog, EventWarning, EventError:
		return s, nil

	case EventCancellation:
		s = terminate(s, ResultCancelled, e.Err, StateCancelled, e.At)
		return s, []Command{{Kind: CommandCancelProvider}}

	case EventTimeout:
		s = terminate(s, ResultTimeout, e.Err, StateTimedOut, e.At)
		return s, []Command{{Kind: CommandCancelProvider}}

	case EventResult:
		s.HasProviderResult = true
		status := e.Result.Status
		if status == "" {
			status = ResultCompleted
		}
		phase := StateCompleted
		switch status {
		case ResultCompleted:
		case ResultFailed, ResultBlocked, ResultAborted:
			phase = StateFailed
		case ResultCancelled:
			phase = StateCancelled
		case ResultTimeout:
			phase = StateTimedOut
		}
		s = terminate(s, status, e.Result.Error, phase, e.At)
		if e.Result.Output != "" {
			s.Result.Output = e.Result.Output
		} else {
			s.Result.Output = string(s.Output)
		}
		return s, []Command{{Kind: CommandFlushEvents}}

	case EventProcessExit:
		if e.ExitCode != 0 && !s.HasProviderResult {
			s = terminate(s, ResultFailed, e.Err, StateFailed, e.At)
		}
		return s, nil
	}

	return s, nil
}

func terminate(s State, status ResultStatus, errMsg string, phase RunState, at time.Time) State {
	s.Phase = phase
	s.Terminal = true
	s.Result.Status = status
	if errMsg != "" {
		s.Result.Error = errMsg
	}
	if s.Result.SessionID == "" {
		s.Result.SessionID = s.SessionID
	}
	s.Result.Usage = s.Usage
	if !at.IsZero() {
		s.Result.FinishedAt = at
	}
	return s
}
