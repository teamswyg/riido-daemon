package main

import "github.com/teamswyg/riido-daemon/internal/taskdb"

func printTaskEvidenceResult(
	taskDBPath string,
	evidence taskdb.TaskEvidenceRecord,
	receipt taskdb.TaskCommandReceiptRecord,
) error {
	return printJSON(struct {
		OK         bool                            `json:"ok"`
		TaskDBPath string                          `json:"task_db_path"`
		Evidence   taskdb.TaskEvidenceRecord       `json:"evidence"`
		Receipt    taskdb.TaskCommandReceiptRecord `json:"receipt"`
	}{
		OK:         true,
		TaskDBPath: taskDBPath,
		Evidence:   evidence,
		Receipt:    receipt,
	})
}
