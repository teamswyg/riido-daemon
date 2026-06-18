package taskdb

func validateReplayedTransition(commandID string, transition TaskTransitionRecord, input TaskTransitionInput, actor, source string) error {
	if transition.TaskID != input.TaskID {
		return commandReplayMismatch(commandID, "task_id")
	}
	if transition.ToState != input.ToState {
		return commandReplayMismatch(commandID, "to_state")
	}
	if transition.EventType != input.Event {
		return commandReplayMismatch(commandID, "event_type")
	}
	if transition.Actor != actor {
		return commandReplayMismatch(commandID, "actor")
	}
	if transition.Source != source {
		return commandReplayMismatch(commandID, "source")
	}
	if transition.Reason != input.Reason {
		return commandReplayMismatch(commandID, "reason")
	}
	return nil
}
