package claude

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func claudeMessageContent(raw agentbridge.RawEvent) []map[string]any {
	message, _ := raw.Payload["message"].(map[string]any)
	content, _ := message["content"].([]any)
	out := make([]map[string]any, 0, len(content))
	for _, item := range content {
		obj, ok := item.(map[string]any)
		if ok {
			out = append(out, obj)
		}
	}
	return out
}
