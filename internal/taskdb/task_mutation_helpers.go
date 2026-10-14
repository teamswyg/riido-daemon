package taskdb

import "sort"

func findTaskIndex(db TaskDB, taskID string) int {
	for i, record := range db.Tasks {
		if record.ID == taskID {
			return i
		}
	}
	return -1
}

func markTaskUpdated(db *TaskDB, index int, stamp string) {
	db.Tasks[index].UpdatedAt = stamp
	db.UpdatedAt = stamp
}

func finalizeTaskMutation(db *TaskDB) {
	recountTransitions(db)
	recountEvidence(db)
	recountCommandReceipts(db)
	sort.Slice(db.Tasks, func(i, j int) bool {
		return db.Tasks[i].ID < db.Tasks[j].ID
	})
}
