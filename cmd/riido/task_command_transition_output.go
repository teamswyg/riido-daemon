package main

import "github.com/teamswyg/riido-daemon/internal/taskdb"

func printTaskTransitionResult(
	taskDBPath string,
	transition taskdb.TaskTransitionRecord,
	receipt taskdb.TaskCommandReceiptRecord,
) error {
	return printJSON(struct {
		OK         bool                            `json:"ok"`
		TaskDBPath string                          `json:"task_db_path"`
		Transition taskdb.TaskTransitionRecord     `json:"transition"`
		Receipt    taskdb.TaskCommandReceiptRecord `json:"receipt"`
	}{
		OK:         true,
		TaskDBPath: taskDBPath,
		Transition: transition,
		Receipt:    receipt,
	})
}
