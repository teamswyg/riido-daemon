// Package history selects externally authored turns of an already bound Riido session.
package history

import (
	"errors"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const assignmentPrefix = "# Riido AI Agent Assignment\n"

func Enabled(req agentbridge.StartRequest) bool {
	return req.ResumeSessionID != "" && strings.HasPrefix(req.Prompt, assignmentPrefix)
}

// Riido-owned prompt/response pairs are already projected by the server. Keep
// external user/assistant messages after the first verified task anchor only.
func External(req agentbridge.StartRequest, messages []agentbridge.NativeHistoryMessage) ([]agentbridge.Event, error) {
	taskID := req.Metadata["task_id"]
	if taskID == "" {
		return nil, errors.New("native history: missing logical task identity")
	}
	anchored, owned := false, false
	totalBytes := 0
	seen := make(map[string]bool)
	var out []agentbridge.Event
	for i, message := range messages {
		if message.SessionID != req.ResumeSessionID {
			return nil, errors.New("native history: session identity mismatch")
		}
		if message.Role == "user" && strings.TrimSpace(message.Body) != "" {
			owned = strings.HasPrefix(message.Body, assignmentPrefix)
			if owned {
				if !strings.Contains(message.Body, "- task_id: "+taskID+"\n") {
					return nil, errors.New("native history: task identity mismatch")
				}
				anchored = true
			}
		}
		if !anchored || owned || (message.Role != "user" && message.Role != "agent") || strings.TrimSpace(message.Body) == "" {
			continue
		}
		if message.MessageID == "" || len(message.MessageID) > 1024 || len(message.Body) > 65536 || i >= 1000000 {
			return nil, errors.New("native history: invalid message identity or oversized text")
		}
		if seen[message.MessageID] {
			continue
		}
		seen[message.MessageID] = true
		message.Ordinal = i + 1
		totalBytes += len(message.Body)
		if len(out) >= 2000 || totalBytes > 4<<20 {
			return nil, errors.New("native history: import exceeds 2000 messages or 4 MiB of visible text")
		}
		out = append(out, agentbridge.Event{Kind: agentbridge.EventProgress, NativeHistory: &message})
	}
	if !anchored {
		return nil, errors.New("native history: Riido task anchor not found; refusing an incomplete import")
	}
	return out, nil
}
