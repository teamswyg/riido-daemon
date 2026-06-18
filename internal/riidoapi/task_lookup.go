package riidoapi

import "github.com/teamswyg/riido-daemon/internal/taskdb"

func findTask(db taskdb.TaskDB, id string) (taskdb.TaskRecord, bool) {
	for _, record := range db.Tasks {
		if record.ID == id {
			return record, true
		}
	}
	return taskdb.TaskRecord{}, false
}
