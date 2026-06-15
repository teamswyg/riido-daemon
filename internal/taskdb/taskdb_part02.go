package taskdb

import (
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func ApplyTaskTransition(existing TaskDB, taskID string, to task.TaskState, event ir.EventType, actor, source, reason string, now time.Time) (TaskDB, TaskTransitionRecord, error) {
	updated, transition, _, err := ApplyGuardedTaskTransition(existing, TaskTransitionInput{
		TaskID:  taskID,
		ToState: to,
		Event:   event,
		Actor:   actor,
		Source:  source,
		Reason:  reason,
		Guard: TaskMutationGuardInput{
			ApprovalID: "approval.riido.legacy",
		},
	}, now)
	return updated, transition, err
}

func ApplyGuardedTaskTransition(existing TaskDB, input TaskTransitionInput, now time.Time) (TaskDB, TaskTransitionRecord, TaskCommandReceiptRecord, error) {
	if input.TaskID == "" {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBInput, "transition.validate", "task id is empty")
	}
	if !input.ToState.Code().IsKnown() {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBState, "transition.validate", "unknown target state: %s", input.ToState)
	}
	if !input.Event.Code().IsTransition() {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBState, "transition.validate", "event %q is not a transition event", input.Event)
	}
	db := normalizeTaskDB(existing)
	index := -1
	for i, record := range db.Tasks {
		if record.ID == input.TaskID {
			index = i
			break
		}
	}
	if index < 0 {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBState, "transition.find-task", "task %s not found", input.TaskID)
	}
	actor := textutil.FirstNonEmpty(input.Actor, "human")
	source := textutil.FirstNonEmpty(input.Source, "riido-cli")
	replayedTransition, replayedReceipt, replayed, err := replayExistingTaskTransition(db, input, actor, source)
	if err != nil {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, err
	}
	if replayed {
		return db, replayedTransition, replayedReceipt, nil
	}
	from := db.Tasks[index].State
	if !task.ValidateTransitionCode(from.Code(), input.ToState.Code(), input.Event.Code()) {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBState, "transition.apply", "illegal task transition: %s --%s--> %s", from, input.Event, input.ToState)
	}
	stamp := timestamp(now)
	receipt, err := buildTaskCommandReceipt(db, db.Tasks[index], "transition", actor, source, input.Guard, now, len(db.CommandReceipts)+1)
	if err != nil {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, err
	}
	transition := TaskTransitionRecord{
		ID:               transitionID(input.TaskID, input.Event, now, len(db.Transitions)+1),
		TaskID:           input.TaskID,
		FromState:        from,
		ToState:          input.ToState,
		EventType:        input.Event,
		Actor:            actor,
		Source:           source,
		Reason:           input.Reason,
		CommandReceiptID: receipt.ID,
		RecordedAt:       stamp,
	}
	receipt.TransitionID = transition.ID
	db.Transitions = append(db.Transitions, transition)
	db.CommandReceipts = append(db.CommandReceipts, receipt)
	db.Tasks[index].State = input.ToState
	db.Tasks[index].UpdatedAt = stamp
	db.UpdatedAt = stamp
	recountTransitions(&db)
	recountEvidence(&db)
	recountCommandReceipts(&db)
	sort.Slice(db.Tasks, func(i, j int) bool {
		return db.Tasks[i].ID < db.Tasks[j].ID
	})
	return db, transition, receipt, nil
}

func AddTaskEvidence(existing TaskDB, input TaskEvidenceInput, now time.Time) (TaskDB, TaskEvidenceRecord, error) {
	updated, evidence, _, err := AddGuardedTaskEvidence(existing, input, now)
	return updated, evidence, err
}

func AddGuardedTaskEvidence(existing TaskDB, input TaskEvidenceInput, now time.Time) (TaskDB, TaskEvidenceRecord, TaskCommandReceiptRecord, error) {
	if input.TaskID == "" {
		return TaskDB{}, TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBInput, "evidence.validate", "task id is empty")
	}
	if strings.TrimSpace(input.Command) == "" {
		return TaskDB{}, TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBInput, "evidence.validate", "evidence command is empty")
	}
	db := normalizeTaskDB(existing)
	index := -1
	for i, record := range db.Tasks {
		if record.ID == input.TaskID {
			index = i
			break
		}
	}
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
	evidenceResult := normalizeEvidenceResult(input.Result, input.ExitCode)
	validationGate := textutil.FirstNonEmpty(input.ValidationGate, TaskEvidenceValidationV1)
	providerRunID := textutil.FirstNonEmpty(input.ProviderRunID, receipt.CommandID)
	providerRunResult := textutil.FirstNonEmpty(input.ProviderRunResult, evidenceResult)
	evidence := TaskEvidenceRecord{
		ID:                evidenceID(input.TaskID, now, len(db.Evidence)+1),
		TaskID:            input.TaskID,
		ProjectID:         taskRecord.ProjectID,
		DocumentID:        taskRecord.SourceDocumentID,
		DocumentPath:      taskRecord.SourceDocumentPath,
		Command:           strings.TrimSpace(input.Command),
		ExitCode:          input.ExitCode,
		Result:            evidenceResult,
		ValidationGate:    validationGate,
		ProviderRunID:     providerRunID,
		ProviderRunResult: providerRunResult,
		Actor:             actor,
		Source:            source,
		Summary:           input.Summary,
		CommandReceiptID:  receipt.ID,
		RecordedAt:        stamp,
	}
	receipt.EvidenceID = evidence.ID
	receipt.Result = evidence.Result
	db.Evidence = append(db.Evidence, evidence)
	db.CommandReceipts = append(db.CommandReceipts, receipt)
	db.Tasks[index].UpdatedAt = stamp
	db.UpdatedAt = stamp
	recountTransitions(&db)
	recountEvidence(&db)
	recountCommandReceipts(&db)
	sort.Slice(db.Tasks, func(i, j int) bool {
		return db.Tasks[i].ID < db.Tasks[j].ID
	})
	return db, evidence, receipt, nil
}

func ParseTaskState(value string) (task.TaskState, error) {
	code := task.ParseTaskStateCode(value)
	if !code.IsKnown() {
		return "", taskDBErrorf(ErrTaskDBState, "parse-state", "unknown task state: %s", value)
	}
	return code.TaskState(), nil
}
