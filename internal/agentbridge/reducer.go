package agentbridge

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
	if next, cmds, handled := reduceTerminalState(s, e); handled {
		return next, cmds
	}
	s = recordSemanticActivity(s, e)
	switch e.Kind {
	case EventSessionIdentified:
		return reduceSessionIdentified(s, e)
	case EventLifecycle:
		return reduceLifecycle(s, e)
	case EventToolApprovalNeeded:
		return reduceToolApprovalNeeded(s, e, approver)
	case EventToolCallStarted:
		return reduceToolCallStarted(s, e)
	case EventToolCallDelta:
		return s, nil
	case EventToolCallCompleted, EventToolCallFailed:
		return reduceToolCallFinished(s, e)
	case EventUsageDelta:
		return reduceUsageDelta(s, e)
	case EventTextDelta:
		return reduceTextDelta(s, e)
	case EventThinkingDelta, EventProgress, EventLog, EventWarning, EventError:
		return s, nil
	case EventCancellation:
		return reduceCancellation(s, e)
	case EventTimeout:
		return reduceTimeout(s, e)
	case EventResult:
		return reduceResult(s, e)
	case EventProcessExit:
		return reduceProcessExit(s, e)
	default:
		return s, nil
	}
}
