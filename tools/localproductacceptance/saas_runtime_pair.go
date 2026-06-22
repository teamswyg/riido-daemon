package main

func choosePreparedRuntimePair(runtimes []preparedRuntime) ([2]preparedRuntime, bool) {
	for _, kind := range []string{"codex", "claude_code", "openclaw", "cursor"} {
		if pair, ok := pairForRuntimeKind(runtimes, kind); ok {
			return pair, true
		}
	}
	return [2]preparedRuntime{}, false
}

func pairForRuntimeKind(runtimes []preparedRuntime, kind string) ([2]preparedRuntime, bool) {
	for i, first := range runtimes {
		for _, second := range runtimes[i+1:] {
			if first.Kind == kind && second.Kind == kind && first.RuntimeID != second.RuntimeID {
				return [2]preparedRuntime{first, second}, true
			}
		}
	}
	return [2]preparedRuntime{}, false
}
