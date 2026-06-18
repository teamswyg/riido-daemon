package riidoapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func (s Server) applyTransition(params json.RawMessage) (TransitionResponse, error) {
	var req TransitionRequest
	if len(params) == 0 {
		return TransitionResponse{}, errors.New("transition params are required")
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return TransitionResponse{}, fmt.Errorf("decode transition params: %w", err)
	}
	updated, transition, receipt, err := s.transitionTask(req)
	if err != nil {
		return TransitionResponse{}, err
	}
	record, ok := findTask(updated, req.TaskID)
	if !ok {
		return TransitionResponse{}, fmt.Errorf("task %s not found after transition", req.TaskID)
	}
	return TransitionResponse{TaskDBPath: s.config.TaskDBPath, Task: record, Transition: transition, Receipt: receipt}, nil
}

func (s Server) transitionTask(req TransitionRequest) (taskdb.TaskDB, taskdb.TaskTransitionRecord, taskdb.TaskCommandReceiptRecord, error) {
	to, err := taskdb.ParseTaskState(req.ToState)
	if err != nil {
		return taskdb.TaskDB{}, taskdb.TaskTransitionRecord{}, taskdb.TaskCommandReceiptRecord{}, err
	}
	db, err := taskdb.LoadTaskDB(s.config.TaskDBPath)
	if err != nil {
		return taskdb.TaskDB{}, taskdb.TaskTransitionRecord{}, taskdb.TaskCommandReceiptRecord{}, err
	}
	updated, transition, receipt, err := taskdb.ApplyGuardedTaskTransition(db, transitionInput(req, to), time.Now())
	if err != nil {
		return taskdb.TaskDB{}, taskdb.TaskTransitionRecord{}, taskdb.TaskCommandReceiptRecord{}, err
	}
	if err := taskdb.SaveTaskDB(s.config.TaskDBPath, updated); err != nil {
		return taskdb.TaskDB{}, taskdb.TaskTransitionRecord{}, taskdb.TaskCommandReceiptRecord{}, err
	}
	return updated, transition, receipt, nil
}

func transitionInput(req TransitionRequest, to task.TaskState) taskdb.TaskTransitionInput {
	return taskdb.TaskTransitionInput{
		TaskID:  req.TaskID,
		ToState: to,
		Event:   ir.EventType(req.EventType),
		Actor:   req.Actor,
		Source:  req.Source,
		Reason:  req.Reason,
		Guard: taskdb.TaskMutationGuardInput{
			CommandID: req.CommandID, Provider: req.Provider, DecisionLLM: req.DecisionLLM, ApprovalID: req.ApprovalID,
		},
	}
}
