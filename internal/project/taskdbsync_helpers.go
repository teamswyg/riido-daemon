package project

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func recountCommandReceipts(db *taskdb.TaskDB) {
	counts := make(map[string]int, len(db.Tasks))
	for _, receipt := range db.CommandReceipts {
		counts[receipt.TaskID]++
	}
	for index := range db.Tasks {
		db.Tasks[index].CommandReceiptCount = counts[db.Tasks[index].ID]
	}
}

func recountTransitions(db *taskdb.TaskDB) {
	counts := make(map[string]int, len(db.Tasks))
	for _, transition := range db.Transitions {
		counts[transition.TaskID]++
	}
	for index := range db.Tasks {
		db.Tasks[index].TransitionCount = counts[db.Tasks[index].ID]
	}
}

func recountEvidence(db *taskdb.TaskDB) {
	counts := make(map[string]int, len(db.Tasks))
	for _, evidence := range db.Evidence {
		counts[evidence.TaskID]++
	}
	for index := range db.Tasks {
		db.Tasks[index].EvidenceCount = counts[db.Tasks[index].ID]
	}
}

func hasTransition(transitions []taskdb.TaskTransitionRecord, id string) bool {
	for _, transition := range transitions {
		if transition.ID == id {
			return true
		}
	}
	return false
}

func initialTaskTransitionID(taskID string) string {
	return "transition:" + taskID + ":created"
}

func timestamp(now time.Time) string {
	return now.UTC().Format(time.RFC3339Nano)
}
