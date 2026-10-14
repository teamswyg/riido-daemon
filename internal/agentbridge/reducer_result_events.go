package agentbridge

func reduceCancellation(s State, e Event) (State, []Command) {
	s = terminate(s, ResultCancelled, e.Err, StateCancelled, e.At)
	return s, []Command{{Kind: CommandCancelProvider}}
}

func reduceTimeout(s State, e Event) (State, []Command) {
	s = terminate(s, ResultTimeout, e.Err, StateTimedOut, e.At)
	return s, []Command{{Kind: CommandCancelProvider}}
}

func reduceResult(s State, e Event) (State, []Command) {
	s.HasProviderResult = true
	status := normalizeResultStatus(e.Result.Status)
	s = terminate(s, status, e.Result.Error, phaseForResultStatus(status), e.At)
	if e.Result.Output != "" {
		s.Result.Output = e.Result.Output
	} else {
		s.Result.Output = string(s.Output)
	}
	return s, []Command{{Kind: CommandFlushEvents}}
}

func reduceProcessExit(s State, e Event) (State, []Command) {
	if e.ExitCode != 0 && !s.HasProviderResult {
		s = terminate(s, ResultFailed, e.Err, StateFailed, e.At)
	}
	return s, nil
}

func normalizeResultStatus(status ResultStatus) ResultStatus {
	if status == "" {
		return ResultCompleted
	}
	return status
}

func phaseForResultStatus(status ResultStatus) RunState {
	switch status {
	case ResultFailed, ResultBlocked, ResultAborted:
		return StateFailed
	case ResultCancelled:
		return StateCancelled
	case ResultTimeout:
		return StateTimedOut
	default:
		return StateCompleted
	}
}
