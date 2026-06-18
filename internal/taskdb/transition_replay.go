package taskdb

func replayExistingTaskTransition(db TaskDB, input TaskTransitionInput, actor, source string) (TaskTransitionRecord, TaskCommandReceiptRecord, bool, error) {
	receipt, found, err := findCommandReceiptByCommandID(db, input.Guard.CommandID)
	if err != nil || !found {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, found, err
	}
	if err := validateCommandReceiptReplay(receipt, "transition", input.TaskID, actor, source, input.Guard); err != nil {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, true, err
	}
	transition, err := findReplayedTransition(db, receipt)
	if err != nil {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, true, err
	}
	if err := validateReplayedTransition(receipt.CommandID, transition, input, actor, source); err != nil {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, true, err
	}
	return transition, receipt, true, nil
}

func findReplayedTransition(db TaskDB, receipt TaskCommandReceiptRecord) (TaskTransitionRecord, error) {
	if receipt.TransitionID == "" {
		return TaskTransitionRecord{}, taskDBErrorf(ErrTaskDBReplay, "transition.replay", "command_id %s replay cannot find linked transition id", receipt.CommandID)
	}
	transition, ok := findTransitionByID(db.Transitions, receipt.TransitionID)
	if !ok {
		return TaskTransitionRecord{}, taskDBErrorf(ErrTaskDBReplay, "transition.replay", "command_id %s replay cannot find transition %s", receipt.CommandID, receipt.TransitionID)
	}
	return transition, nil
}
