package taskdb

import (
	"strings"

	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func validateReplayedEvidence(receipt TaskCommandReceiptRecord, evidence TaskEvidenceRecord, input TaskEvidenceInput, actor, source string) error {
	if err := validateReplayedEvidenceCore(receipt.CommandID, evidence, input); err != nil {
		return err
	}
	if err := validateReplayedEvidenceRuntimeFields(receipt, evidence, input); err != nil {
		return err
	}
	return validateReplayedEvidenceActorFields(receipt.CommandID, evidence, input, actor, source)
}

func validateReplayedEvidenceCore(commandID string, evidence TaskEvidenceRecord, input TaskEvidenceInput) error {
	if evidence.TaskID != input.TaskID {
		return commandReplayMismatch(commandID, "task_id")
	}
	if evidence.Command != strings.TrimSpace(input.Command) {
		return commandReplayMismatch(commandID, "command")
	}
	if evidence.ExitCode != input.ExitCode {
		return commandReplayMismatch(commandID, "exit_code")
	}
	if evidence.Result != normalizeEvidenceResult(input.Result, input.ExitCode) {
		return commandReplayMismatch(commandID, "result")
	}
	return nil
}

func validateReplayedEvidenceRuntimeFields(receipt TaskCommandReceiptRecord, evidence TaskEvidenceRecord, input TaskEvidenceInput) error {
	validationGate := textutil.FirstNonEmpty(input.ValidationGate, TaskEvidenceValidationV1)
	providerRunID := textutil.FirstNonEmpty(input.ProviderRunID, receipt.CommandID)
	providerRunResult := textutil.FirstNonEmpty(input.ProviderRunResult, evidence.Result)
	if !replayStringFieldMatches(evidence.ValidationGate, validationGate, input.ValidationGate != "") {
		return commandReplayMismatch(receipt.CommandID, "validation_gate")
	}
	if !replayStringFieldMatches(evidence.ProviderRunID, providerRunID, input.ProviderRunID != "") {
		return commandReplayMismatch(receipt.CommandID, "provider_run_id")
	}
	if !replayStringFieldMatches(evidence.ProviderRunResult, providerRunResult, input.ProviderRunResult != "") {
		return commandReplayMismatch(receipt.CommandID, "provider_run_result")
	}
	return nil
}
