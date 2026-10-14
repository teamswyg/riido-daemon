package taskdb

import "strings"

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
	if err := validateReceiptReplayGuard(receipt, guard); err != nil {
		return err
	}
	if receipt.GuardDecision != "accepted" {
		return taskDBErrorf(ErrTaskDBReplay, "receipt.replay", "command_id %s replay cannot reuse receipt with guard decision %s", receipt.CommandID, receipt.GuardDecision)
	}
	return nil
}

func validateReceiptReplayGuard(receipt TaskCommandReceiptRecord, guard TaskMutationGuardInput) error {
	if receipt.ApprovalID != strings.TrimSpace(guard.ApprovalID) {
		return commandReplayMismatch(receipt.CommandID, "approval_id")
	}
	if strings.TrimSpace(guard.Provider) != "" && receipt.Provider != strings.TrimSpace(guard.Provider) {
		return commandReplayMismatch(receipt.CommandID, "provider")
	}
	if strings.TrimSpace(guard.DecisionLLM) != "" && receipt.DecisionLLM != strings.TrimSpace(guard.DecisionLLM) {
		return commandReplayMismatch(receipt.CommandID, "decision_llm")
	}
	return nil
}
