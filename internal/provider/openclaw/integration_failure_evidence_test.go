package openclaw

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func openClawFailureEvidence(events []agentbridge.Event) string {
	var parts []string
	for _, ev := range events {
		if text := evidenceText(ev); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(tailEvidence(parts, 5), " | ")
}

func evidenceText(ev agentbridge.Event) string {
	switch ev.Kind {
	case agentbridge.EventError, agentbridge.EventLog, agentbridge.EventWarning:
		return strings.TrimSpace(firstNonEmpty(ev.Err, ev.Text))
	default:
		return ""
	}
}

func tailEvidence(parts []string, limit int) []string {
	if len(parts) <= limit {
		return parts
	}
	return parts[len(parts)-limit:]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
