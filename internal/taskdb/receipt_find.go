package taskdb

import "strings"

func findCommandReceiptByCommandID(db TaskDB, commandID string) (TaskCommandReceiptRecord, bool, error) {
	commandID = strings.TrimSpace(commandID)
	if commandID == "" {
		return TaskCommandReceiptRecord{}, false, nil
	}
	var found TaskCommandReceiptRecord
	hasFound := false
	for _, receipt := range db.CommandReceipts {
		if receipt.CommandID != commandID {
			continue
		}
		if hasFound {
			return TaskCommandReceiptRecord{}, true, taskDBErrorf(ErrTaskDBReplay, "receipt.find-by-command-id", "command_id %s is not unique in task DB", commandID)
		}
		found = receipt
		hasFound = true
	}
	return found, hasFound, nil
}
