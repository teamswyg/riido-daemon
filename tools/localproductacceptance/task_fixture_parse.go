package main

func firstTeamID(payload map[string]any) string {
	for _, key := range []string{"teams", "data", "items"} {
		if id := firstArrayObjectString(payload[key], "teamId", "team_id", "id"); id != "" {
			return id
		}
	}
	return firstString(payload, "teamId", "team_id", "id")
}

func firstArrayObjectString(value any, keys ...string) string {
	items, _ := value.([]any)
	for _, item := range items {
		object, _ := item.(map[string]any)
		if id := firstString(object, keys...); id != "" {
			return id
		}
	}
	return ""
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if text := stringValue(payload[key]); text != "" {
			return text
		}
	}
	return ""
}
