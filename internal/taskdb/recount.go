package taskdb

func recountTaskDB(db *TaskDB) {
	recountTransitions(db)
	recountEvidence(db)
	recountCommandReceipts(db)
}

func recountTransitions(db *TaskDB) {
	counts := make(map[string]int, len(db.Tasks))
	for _, transition := range db.Transitions {
		counts[transition.TaskID]++
	}
	for index := range db.Tasks {
		db.Tasks[index].TransitionCount = counts[db.Tasks[index].ID]
	}
}

func recountEvidence(db *TaskDB) {
	counts := make(map[string]int, len(db.Tasks))
	for _, evidence := range db.Evidence {
		counts[evidence.TaskID]++
	}
	for index := range db.Tasks {
		db.Tasks[index].EvidenceCount = counts[db.Tasks[index].ID]
	}
}

func recountCommandReceipts(db *TaskDB) {
	counts := make(map[string]int, len(db.Tasks))
	for _, receipt := range db.CommandReceipts {
		counts[receipt.TaskID]++
	}
	for index := range db.Tasks {
		db.Tasks[index].CommandReceiptCount = counts[db.Tasks[index].ID]
	}
}
