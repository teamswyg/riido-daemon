package main

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func runtimeKinds() map[string]bool {
	out := map[string]bool{}
	for _, kind := range agentbridge.EventKinds() {
		out[string(kind)] = kind.IsSemanticActivity()
	}
	return out
}

func category(semantic bool) string {
	if semantic {
		return "semantic"
	}
	return "non_semantic"
}
