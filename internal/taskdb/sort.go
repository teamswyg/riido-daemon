package taskdb

import "sort"

func sortTaskDB(db *TaskDB) {
	sort.Slice(db.ProviderCandidates, func(i, j int) bool {
		return db.ProviderCandidates[i].ID < db.ProviderCandidates[j].ID
	})
	sort.Slice(db.Tasks, func(i, j int) bool {
		return db.Tasks[i].ID < db.Tasks[j].ID
	})
	sortTimedRecords(db.Transitions, transitionSortKey)
	sortTimedRecords(db.Evidence, evidenceSortKey)
	sortTimedRecords(db.CommandReceipts, receiptSortKey)
}

func sortTimedRecords[T any](records []T, key func(T) (string, string)) {
	sort.Slice(records, func(i, j int) bool {
		leftTime, leftID := key(records[i])
		rightTime, rightID := key(records[j])
		return leftTime < rightTime || leftTime == rightTime && leftID < rightID
	})
}
