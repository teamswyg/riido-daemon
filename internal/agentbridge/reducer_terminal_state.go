package agentbridge

import "time"

func reduceTerminalState(s State, e Event) (State, []Command, bool) {
	if !s.Terminal {
		return s, nil, false
	}
	// Invariant 2: usage may still accumulate; phase stays put.
	if e.Kind == EventUsageDelta {
		s.Usage = s.Usage.Add(e.Usage)
	}
	return s, nil, true
}

func recordSemanticActivity(s State, e Event) State {
	if e.Kind.IsSemanticActivity() && !e.At.IsZero() {
		s.LastSemanticActivity = e.At
	}
	return s
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
