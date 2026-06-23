package main

import "strings"

func firstTaskID(value any) string {
	switch typed := value.(type) {
	case map[string]any:
		if isTaskRecord(typed) {
			if id := firstString(typed, "componentId", "component_id", "id"); id != "" {
				return id
			}
		}
		for _, child := range typed {
			if id := firstTaskID(child); id != "" {
				return id
			}
		}
	case []any:
		for _, child := range typed {
			if id := firstTaskID(child); id != "" {
				return id
			}
		}
	}
	return ""
}

func isTaskRecord(record map[string]any) bool {
	text := strings.ToLower(firstString(record, "componentType", "component_type", "type"))
	return text == "task"
}
