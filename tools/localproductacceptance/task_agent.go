package main

type taskAgentCandidate struct {
	AgentID     string
	RuntimeKind string
	RuntimeID   string
	Name        string
}

func taskAgentCandidates(payload map[string]any) []taskAgentCandidate {
	raw, _ := payload["agents"].([]any)
	out := make([]taskAgentCandidate, 0, len(raw))
	for _, item := range raw {
		agent, _ := item.(map[string]any)
		id, _ := agent["agent_id"].(string)
		if id == "" {
			continue
		}
		out = append(out, taskAgentCandidate{
			AgentID:     id,
			RuntimeKind: stringValue(agent["runtime_kind"]),
			RuntimeID:   stringValue(agent["runtime_id"]),
			Name:        stringValue(agent["name"]),
		})
	}
	return out
}

func chooseTaskAgentPair(agents []taskAgentCandidate) (taskAgentPair, bool) {
	if pair, ok := sameRuntimeKindPair(agents); ok {
		return pair, true
	}
	if len(agents) < 2 {
		return taskAgentPair{}, false
	}
	return taskAgentPair{First: agents[0], Second: agents[1]}, true
}

func sameRuntimeKindPair(agents []taskAgentCandidate) (taskAgentPair, bool) {
	for i, first := range agents {
		for _, second := range agents[i+1:] {
			if first.RuntimeKind != "" && first.RuntimeKind == second.RuntimeKind {
				return taskAgentPair{First: first, Second: second}, true
			}
		}
	}
	return taskAgentPair{}, false
}

func stringValue(value any) string {
	out, _ := value.(string)
	return out
}
