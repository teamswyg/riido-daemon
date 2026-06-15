package taskdb

import (
	"strings"

	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func replayExistingTaskEvidence(db TaskDB, input TaskEvidenceInput, actor, source string) (TaskEvidenceRecord, TaskCommandReceiptRecord, bool, error) {
	receipt, found, err := findCommandReceiptByCommandID(db, input.Guard.CommandID)
	if err != nil || !found {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, found, err
	}
	if err := validateCommandReceiptReplay(receipt, "evidence", input.TaskID, actor, source, input.Guard); err != nil {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, err
	}
	if receipt.EvidenceID == "" {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, taskDBErrorf(ErrTaskDBReplay, "evidence.replay", "command_id %s replay cannot find linked evidence id", receipt.CommandID)
	}
	evidence, ok := findEvidenceByID(db.Evidence, receipt.EvidenceID)
	if !ok {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, taskDBErrorf(ErrTaskDBReplay, "evidence.replay", "command_id %s replay cannot find evidence %s", receipt.CommandID, receipt.EvidenceID)
	}
	if evidence.TaskID != input.TaskID {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "task_id")
	}
	if evidence.Command != strings.TrimSpace(input.Command) {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "command")
	}
	if evidence.ExitCode != input.ExitCode {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "exit_code")
	}
	if evidence.Result != normalizeEvidenceResult(input.Result, input.ExitCode) {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "result")
	}
	validationGate := textutil.FirstNonEmpty(input.ValidationGate, TaskEvidenceValidationV1)
	providerRunID := textutil.FirstNonEmpty(input.ProviderRunID, receipt.CommandID)
	providerRunResult := textutil.FirstNonEmpty(input.ProviderRunResult, evidence.Result)
	if !replayStringFieldMatches(evidence.ValidationGate, validationGate, input.ValidationGate != "") {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "validation_gate")
	}
	if !replayStringFieldMatches(evidence.ProviderRunID, providerRunID, input.ProviderRunID != "") {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "provider_run_id")
	}
	if !replayStringFieldMatches(evidence.ProviderRunResult, providerRunResult, input.ProviderRunResult != "") {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "provider_run_result")
	}
	if evidence.Actor != actor {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "actor")
	}
	if evidence.Source != source {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "source")
	}
	if evidence.Summary != input.Summary {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "summary")
	}
	return evidence, receipt, true, nil
}

func validateCommandReceiptReplay(receipt TaskCommandReceiptRecord, kind, taskID, actor, source string, guard TaskMutationGuardInput) error {
	if receipt.Kind != kind {
		return commandReplayMismatch(receipt.CommandID, "kind")
	}
	if receipt.TaskID != taskID {
		return commandReplayMismatch(receipt.CommandID, "task_id")
	}
	if receipt.Actor != actor {
		return commandReplayMismatch(receipt.CommandID, "actor")
	}
	if receipt.Source != source {
		return commandReplayMismatch(receipt.CommandID, "source")
	}
	if receipt.ApprovalID != strings.TrimSpace(guard.ApprovalID) {
		return commandReplayMismatch(receipt.CommandID, "approval_id")
	}
	if strings.TrimSpace(guard.Provider) != "" && receipt.Provider != strings.TrimSpace(guard.Provider) {
		return commandReplayMismatch(receipt.CommandID, "provider")
	}
	if strings.TrimSpace(guard.DecisionLLM) != "" && receipt.DecisionLLM != strings.TrimSpace(guard.DecisionLLM) {
		return commandReplayMismatch(receipt.CommandID, "decision_llm")
	}
	if receipt.GuardDecision != "accepted" {
		return taskDBErrorf(ErrTaskDBReplay, "receipt.replay", "command_id %s replay cannot reuse receipt with guard decision %s", receipt.CommandID, receipt.GuardDecision)
	}
	return nil
}

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

func findTransitionByID(transitions []TaskTransitionRecord, id string) (TaskTransitionRecord, bool) {
	for _, transition := range transitions {
		if transition.ID == id {
			return transition, true
		}
	}
	return TaskTransitionRecord{}, false
}

func findEvidenceByID(evidenceRecords []TaskEvidenceRecord, id string) (TaskEvidenceRecord, bool) {
	for _, evidence := range evidenceRecords {
		if evidence.ID == id {
			return evidence, true
		}
	}
	return TaskEvidenceRecord{}, false
}

func commandReplayMismatch(commandID, field string) error {
	return taskDBErrorf(ErrTaskDBReplay, "receipt.replay", "command_id %s replay mismatch on %s", commandID, field)
}

func replayStringFieldMatches(existing, expected string, required bool) bool {
	if existing == expected {
		return true
	}
	return existing == "" && !required
}
