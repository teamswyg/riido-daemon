package ingest

import (
	"maps"

	"github.com/teamswyg/riido-contracts/ir"
)

func copyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	maps.Copy(out, in)
	return out
}

func fsmVersionForEvent(eventType ir.EventType, source int) int {
	if eventType.IsTransition() {
		return source
	}
	return 0
}
