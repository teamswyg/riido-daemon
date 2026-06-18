package agentbridge

func reduceToolApprovalNeeded(s State, e Event, approver AutoApprover) (State, []Command) {
	s.Phase = StateWaitingToolApproval
	s = recordToolRef(s, e.Tool)
	if approver != nil && approver(e.Tool) {
		return s, []Command{{
			Kind:              CommandApproveTool,
			ToolID:            e.Tool.ID,
			ProviderRequestID: e.Tool.ProviderRequestID,
		}}
	}
	return s, nil
}

func reduceToolCallStarted(s State, e Event) (State, []Command) {
	s.Phase = StateToolRunning
	s = recordToolRef(s, e.Tool)
	return s, nil
}

func reduceToolCallFinished(s State, e Event) (State, []Command) {
	if e.Tool.ID != "" {
		delete(s.Tools, e.Tool.ID)
	}
	if len(s.Tools) == 0 {
		s.Phase = StateRunning
	}
	return s, nil
}

func recordToolRef(s State, tool ToolRef) State {
	if tool.ID != "" {
		s.Tools[tool.ID] = tool
	}
	return s
}
