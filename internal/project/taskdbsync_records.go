package project

import (
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func syncTaskDBRecords(db taskdb.TaskDB, sourceTasks []TaskState, stamp string) taskdb.TaskDB {
	records := make(map[string]taskdb.TaskRecord, len(db.Tasks)+len(sourceTasks))
	for _, record := range db.Tasks {
		records[record.ID] = record
	}
	for _, sourceTask := range sourceTasks {
		record, exists := records[sourceTask.ID]
		if !exists {
			record = newTaskDBRecord(sourceTask.ID, stamp)
			transition := initialTaskDBTransition(sourceTask.ID, stamp)
			if !hasTransition(db.Transitions, transition.ID) {
				db.Transitions = append(db.Transitions, transition)
			}
		}
		records[sourceTask.ID] = applyTaskDBSource(record, sourceTask, stamp)
	}
	db.Tasks = mapToSortedTasks(records)
	return db
}

func newTaskDBRecord(taskID, stamp string) taskdb.TaskRecord {
	return taskdb.TaskRecord{
		ID:        taskID,
		State:     task.StateCreated,
		CreatedAt: stamp,
	}
}

func initialTaskDBTransition(taskID, stamp string) taskdb.TaskTransitionRecord {
	return taskdb.TaskTransitionRecord{
		ID:         initialTaskTransitionID(taskID),
		TaskID:     taskID,
		ToState:    task.StateCreated,
		EventType:  ir.EventTaskCreated,
		Actor:      "riido",
		Source:     "riido.mwsd.sync",
		Reason:     "document task source discovered",
		RecordedAt: stamp,
	}
}
