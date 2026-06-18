package taskdb

import (
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func AddGuardedTaskEvidence(existing TaskDB, input TaskEvidenceInput, now time.Time) (TaskDB, TaskEvidenceRecord, TaskCommandReceiptRecord, error) {
	if input.TaskID == "" {
		return TaskDB{}, TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBInput, "evidence.validate", "task id is empty")
	}
	if strings.TrimSpace(input.Command) == "" {
		return TaskDB{}, TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBInput, "evidence.validate", "evidence command is empty")
	}
	db := normalizeTaskDB(existing)
	index := findTaskIndex(db, input.TaskID)
	if index < 0 {
		return TaskDB{}, TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBState, "evidence.find-task", "task %s not found", input.TaskID)
	}
	stamp := timestamp(now)
	taskRecord := db.Tasks[index]
	actor := textutil.FirstNonEmpty(input.Actor, "daemon")
	source := textutil.FirstNonEmpty(input.Source, "riido-cli")
	replayedEvidence, replayedReceipt, replayed, err := replayExistingTaskEvidence(db, input, actor, source)
	if err != nil {
		return TaskDB{}, TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, err
	}
	if replayed {
		return db, replayedEvidence, replayedReceipt, nil
	}
	receipt, err := buildTaskCommandReceipt(db, taskRecord, "evidence", actor, source, input.Guard, now, len(db.CommandReceipts)+1)
	if err != nil {
		return TaskDB{}, TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, err
	}
	evidence := newTaskEvidenceRecord(input, taskRecord, receipt.ID, receipt.CommandID, actor, source, stamp, now, len(db.Evidence)+1)
	receipt.EvidenceID = evidence.ID
	receipt.Result = evidence.Result
	db.Evidence = append(db.Evidence, evidence)
	db.CommandReceipts = append(db.CommandReceipts, receipt)
	markTaskUpdated(&db, index, stamp)
	finalizeTaskMutation(&db)
	return db, evidence, receipt, nil
}
