package taskdb

func commandReplayMismatch(commandID, field string) error {
	return taskDBErrorf(ErrTaskDBReplay, "receipt.replay", "command_id %s replay mismatch on %s", commandID, field)
}

func replayStringFieldMatches(existing, expected string, required bool) bool {
	if existing == expected {
		return true
	}
	return existing == "" && !required
}
