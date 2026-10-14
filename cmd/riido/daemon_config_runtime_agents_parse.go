package main

import (
	"strings"
	"unicode"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func parseRuntimeAgents(raw string) []runtimeactor.AgentStatus {
	parts := strings.Split(raw, ",")
	out := make([]runtimeactor.AgentStatus, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		out = append(out, runtimeactor.AgentStatus{
			AgentID: slugAgentName(name),
			Name:    name,
			State:   "online",
		})
	}
	return out
}

func slugAgentName(name string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
