package taskdb

import (
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

type TaskTransitionRecord struct {
	ID               string         `json:"id"`
	TaskID           string         `json:"task_id"`
	FromState        task.TaskState `json:"from_state"`
	ToState          task.TaskState `json:"to_state"`
	EventType        ir.EventType   `json:"event_type"`
	Actor            string         `json:"actor"`
	Source           string         `json:"source"`
	Reason           string         `json:"reason"`
	CommandReceiptID string         `json:"command_receipt_id"`
	RecordedAt       string         `json:"recorded_at"`
}
