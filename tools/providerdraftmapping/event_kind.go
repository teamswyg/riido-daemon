package main

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func eventKindByConst() map[string]string {
	out := map[string]string{}
	for _, kind := range agentbridge.EventKinds() {
		out[eventKindConstName(kind)] = string(kind)
	}
	return out
}

func runtimeEventKinds() map[string]struct{} {
	out := map[string]struct{}{}
	for _, kind := range agentbridge.EventKinds() {
		out[string(kind)] = struct{}{}
	}
	return out
}
