package agentbridge

// ToolStartGate decides what to do when a provider reports that a tool call
// has started but no provider approval round-trip is available at that point.
// Nil means do not block started tool calls.
type ToolStartGate func(tool ToolRef) ToolStartDecision

// ToolApprovalGate decides what to do when a provider asks for tool approval
// but the embedding runtime has no human approval round-trip for this run.
// Nil means leave the run waiting for approval.
type ToolApprovalGate func(tool ToolRef) ToolStartDecision

// ToolStartDecision is the provider-neutral session decision for a started
// tool call. The session actor only knows whether to block and what to report;
// C7 owns the policy matrix that produced the decision.
type ToolStartDecision struct {
	Block  bool
	Code   string
	Reason string
}
