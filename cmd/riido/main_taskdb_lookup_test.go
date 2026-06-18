package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func findValidateTestTask(t *testing.T, db taskdb.TaskDB) taskdb.TaskRecord {
	t.Helper()

	for _, record := range db.Tasks {
		if record.ID == validateTestTaskID {
			return record
		}
	}

	t.Fatalf("task %s not found", validateTestTaskID)
	return taskdb.TaskRecord{}
}

func findValidateTestReceipt(
	t *testing.T,
	db taskdb.TaskDB,
	commandID string,
) taskdb.TaskCommandReceiptRecord {
	t.Helper()

	for _, receipt := range db.CommandReceipts {
		if receipt.CommandID == commandID {
			return receipt
		}
	}

	t.Fatalf("receipt for command %s not found", commandID)
	return taskdb.TaskCommandReceiptRecord{}
}
