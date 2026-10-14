package project

import (
	"sort"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func sortTaskDBTransitions(transitions []taskdb.TaskTransitionRecord) {
	sort.Slice(transitions, func(i, j int) bool {
		return transitions[i].RecordedAt < transitions[j].RecordedAt ||
			(transitions[i].RecordedAt == transitions[j].RecordedAt && transitions[i].ID < transitions[j].ID)
	})
}

func sortTaskDBEvidence(evidence []taskdb.TaskEvidenceRecord) {
	sort.Slice(evidence, func(i, j int) bool {
		return evidence[i].RecordedAt < evidence[j].RecordedAt ||
			(evidence[i].RecordedAt == evidence[j].RecordedAt && evidence[i].ID < evidence[j].ID)
	})
}

func sortTaskDBCommandReceipts(receipts []taskdb.TaskCommandReceiptRecord) {
	sort.Slice(receipts, func(i, j int) bool {
		return receipts[i].RecordedAt < receipts[j].RecordedAt ||
			(receipts[i].RecordedAt == receipts[j].RecordedAt && receipts[i].ID < receipts[j].ID)
	})
}

func mapToSortedTasks(records map[string]taskdb.TaskRecord) []taskdb.TaskRecord {
	tasks := make([]taskdb.TaskRecord, 0, len(records))
	for _, record := range records {
		tasks = append(tasks, record)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return tasks
}
