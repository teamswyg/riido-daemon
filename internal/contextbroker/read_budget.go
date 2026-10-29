package contextbroker

import (
	"encoding/json"
	"unicode/utf8"
)

// Captured values are coupled to immutable byte offsets and versions. A local
// shortening cannot issue a cursor for the removed portion of that snapshot.
func capturedReadBudget(view string, data json.RawMessage, budget int) bool {
	if budget <= 0 || utf8.RuneCount(data) <= budget {
		return false
	}
	return capturedReadResponse(view, data)
}

func capturedReadResponse(view string, data json.RawMessage) bool {
	if view == viewContent {
		return true
	}
	var shape struct {
		SchemaVersion string `json:"schema_version"`
		ReadScope     *struct {
			SchemaVersion string `json:"schema_version"`
		} `json:"read_scope"`
		Resource *struct {
			ID string `json:"id"`
		} `json:"resource"`
		Parts []json.RawMessage `json:"parts"`
	}
	if json.Unmarshal(data, &shape) != nil {
		return false
	}
	return shape.SchemaVersion == "riido-read-scope.v1" ||
		(shape.ReadScope != nil && shape.ReadScope.SchemaVersion == "riido-read-scope.v1") ||
		(shape.Resource != nil && shape.Resource.ID != "" && shape.Parts != nil)
}
