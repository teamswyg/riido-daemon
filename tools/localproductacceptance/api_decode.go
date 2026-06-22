package main

import "encoding/json"

func decodeObjectPayload(data []byte) map[string]any {
	var decoded map[string]any
	if json.Unmarshal(data, &decoded) == nil {
		return decoded
	}
	var items []any
	if json.Unmarshal(data, &items) == nil {
		return map[string]any{"items": items}
	}
	return nil
}
