package claude

import "encoding/json"

func marshalClaudeUserFrame(prompt string) ([]byte, error) {
	body, err := json.Marshal(claudeUserFrame(prompt))
	if err != nil {
		return nil, err
	}
	return append(body, '\n'), nil
}

func claudeUserFrame(prompt string) map[string]any {
	return map[string]any{
		"type": "user",
		"message": map[string]any{
			"role": "user",
			"content": []map[string]any{
				{"type": "text", "text": prompt},
			},
		},
	}
}
