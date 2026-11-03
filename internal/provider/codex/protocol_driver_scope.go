package codex

// Identity is pinned by thread/start (or resume) and turn/start responses.
// Legacy notifications without identity remain supported, but explicit foreign
// identities must never mutate this run or reach its reducer/report stream.
func (d *protocolDriver) belongsToRun(p map[string]any) bool {
	if id := threadIDFromParams(p); id != "" && d.threadID != "" && id != d.threadID {
		return false
	}
	if id := turnIDFromParams(p); id != "" && d.turnID != "" && id != d.turnID {
		return false
	}
	return true
}

func turnIDFromParams(p map[string]any) string {
	if id := stringField(p, "turnId"); id != "" {
		return id
	}
	if id := stringField(p, "turn_id"); id != "" {
		return id
	}
	return stringField(mapField(p, "turn"), "id")
}
