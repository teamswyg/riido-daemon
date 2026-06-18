package agentbridge

func reduceSessionIdentified(s State, e Event) (State, []Command) {
	s.SessionID = e.SessionID
	return s, []Command{{Kind: CommandPersistSession}}
}

func reduceLifecycle(s State, e Event) (State, []Command) {
	if e.Phase != "" && !e.Phase.IsTerminal() {
		s.Phase = e.Phase
	}
	return s, nil
}

func reduceUsageDelta(s State, e Event) (State, []Command) {
	s.Usage = s.Usage.Add(e.Usage)
	return s, nil
}

func reduceTextDelta(s State, e Event) (State, []Command) {
	s.Output = append(s.Output, e.Text...)
	return s, nil
}
