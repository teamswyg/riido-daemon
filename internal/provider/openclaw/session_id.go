package openclaw

import (
	"errors"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// ResolveSessionID maps the provider-neutral start request onto
// OpenClaw's mandatory --session-id. ResumeSessionID preserves a
// provider session, while TaskID gives first-run tasks a deterministic,
// provider-safe session id without inventing provider state.
func ResolveSessionID(req agentbridge.StartRequest) (string, error) {
	if strings.TrimSpace(req.ResumeSessionID) != "" {
		return req.ResumeSessionID, nil
	}
	if strings.TrimSpace(req.TaskID) != "" {
		return sessionIDFromTaskID(req.TaskID), nil
	}
	return "", errors.New("openclaw: SessionID is required (set ResumeSessionID or TaskID)")
}

func isOpenClawSessionID(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 && !isOpenClawSessionStartRune(r) {
			return false
		}
		if !isOpenClawSessionRune(r) {
			return false
		}
	}
	return true
}

func isOpenClawSessionStartRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func isOpenClawSessionRune(r rune) bool {
	return isOpenClawSessionStartRune(r) || r == '-' || r == '_'
}
