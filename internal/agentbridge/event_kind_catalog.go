package agentbridge

// EventKinds returns the provider-neutral event taxonomy in stable order.
func EventKinds() []EventKind {
	return []EventKind{
		EventLifecycle,
		EventSessionIdentified,
		EventTextDelta,
		EventThinkingDelta,
		EventToolCallStarted,
		EventToolCallDelta,
		EventToolCallCompleted,
		EventToolCallFailed,
		EventToolApprovalNeeded,
		EventUsageDelta,
		EventProgress,
		EventSemanticActivity,
		EventLog,
		EventWarning,
		EventError,
		EventResult,
		EventCancellation,
		EventTimeout,
		EventProcessExit,
	}
}
