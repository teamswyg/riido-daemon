package taskdb

import (
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func ApplyGuardedTaskTransition(existing TaskDB, input TaskTransitionInput, now time.Time) (TaskDB, TaskTransitionRecord, TaskCommandReceiptRecord, error) {
	if err := validateTaskTransitionInput(input); err != nil {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, err
	}
	db := normalizeTaskDB(existing)
	index := findTaskIndex(db, input.TaskID)
	if index < 0 {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBState, "transition.find-task", "task %s not found", input.TaskID)
	}
	actor := textutil.FirstNonEmpty(input.Actor, "human")
	source := textutil.FirstNonEmpty(input.Source, "riido-cli")
	replayedTransition, replayedReceipt, replayed, err := replayExistingTaskTransition(db, input, actor, source)
	if err != nil {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, err
	}
	if replayed {
		return db, replayedTransition, replayedReceipt, nil
	}
	from := db.Tasks[index].State
	if !task.ValidateTransitionCode(from.Code(), input.ToState.Code(), input.Event.Code()) {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBState, "transition.apply", "illegal task transition: %s --%s--> %s", from, input.Event, input.ToState)
	}
	stamp := timestamp(now)
	receipt, err := buildTaskCommandReceipt(db, db.Tasks[index], "transition", actor, source, input.Guard, now, len(db.CommandReceipts)+1)
	if err != nil {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, err
	}
	transition := newTaskTransitionRecord(input, receipt.ID, from, actor, source, stamp, now, len(db.Transitions)+1)
	receipt.TransitionID = transition.ID
	db.Transitions = append(db.Transitions, transition)
	db.CommandReceipts = append(db.CommandReceipts, receipt)
	db.Tasks[index].State = input.ToState
	markTaskUpdated(&db, index, stamp)
	finalizeTaskMutation(&db)
	return db, transition, receipt, nil
}

func validateTaskTransitionInput(input TaskTransitionInput) error {
	if input.TaskID == "" {
		return taskDBErrorf(ErrTaskDBInput, "transition.validate", "task id is empty")
	}
	if !input.ToState.Code().IsKnown() {
		return taskDBErrorf(ErrTaskDBState, "transition.validate", "unknown target state: %s", input.ToState)
	}
	if !input.Event.Code().IsTransition() {
		return taskDBErrorf(ErrTaskDBState, "transition.validate", "event %q is not a transition event", input.Event)
	}
	return nil
}
