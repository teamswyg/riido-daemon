package taskdb

import (
	"time"

	"github.com/teamswyg/riido-contracts/task"
)

func newTaskTransitionRecord(input TaskTransitionInput, receiptID string, from task.TaskState, actor, source, stamp string, now time.Time, ordinal int) TaskTransitionRecord {
	return TaskTransitionRecord{
		ID:               transitionID(input.TaskID, input.Event, now, ordinal),
		TaskID:           input.TaskID,
		FromState:        from,
		ToState:          input.ToState,
		EventType:        input.Event,
		Actor:            actor,
		Source:           source,
		Reason:           input.Reason,
		CommandReceiptID: receiptID,
		RecordedAt:       stamp,
	}
}
