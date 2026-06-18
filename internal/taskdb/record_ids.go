package taskdb

import (
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func buildTaskCommandReceipt(db TaskDB, taskRecord TaskRecord, kind, actor, source string, guard TaskMutationGuardInput, now time.Time, ordinal int) (TaskCommandReceiptRecord, error) {
	provider := textutil.FirstNonEmpty(guard.Provider, taskRecord.RecommendedProvider)
	provider = textutil.FirstNonEmpty(provider, db.RecommendedProvider)
	decisionLLM := textutil.FirstNonEmpty(guard.DecisionLLM, taskRecord.RecommendedDecisionLLM)
	decisionLLM = textutil.FirstNonEmpty(decisionLLM, db.RecommendedDecisionLLM)
	approvalID := strings.TrimSpace(guard.ApprovalID)
	commandID := strings.TrimSpace(guard.CommandID)
	if commandID == "" {
		commandID = generatedCommandID(taskRecord.ID, kind, now, ordinal)
	}
	requiresHumanApproval := taskRecord.RequiresHumanApproval || db.DecisionGate == "human-approval-required"
	receipt := TaskCommandReceiptRecord{
		ID:                     commandReceiptID(taskRecord.ID, kind, now, ordinal),
		CommandID:              commandID,
		Kind:                   kind,
		TaskID:                 taskRecord.ID,
		Actor:                  actor,
		Source:                 source,
		Provider:               provider,
		DecisionLLM:            decisionLLM,
		ApprovalID:             approvalID,
		DecisionGate:           db.DecisionGate,
		RequiresHumanApproval:  requiresHumanApproval,
		RecommendedProvider:    textutil.FirstNonEmpty(taskRecord.RecommendedProvider, db.RecommendedProvider),
		RecommendedDecisionLLM: textutil.FirstNonEmpty(taskRecord.RecommendedDecisionLLM, db.RecommendedDecisionLLM),
		HarnessNextDirection:   taskRecord.HarnessNextDirection,
		ReplayPolicy:           TaskCommandReplayPolicyV1,
		RecordedAt:             timestamp(now),
	}
	if requiresHumanApproval && approvalID == "" {
		return TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBGuard, "receipt.build", "task %s requires approval_id before %s mutation", taskRecord.ID, kind)
	}
	if provider == "" {
		return TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBGuard, "receipt.build", "task %s has no provider for %s mutation", taskRecord.ID, kind)
	}
	if !providerCandidateAvailable(db.ProviderCandidates, provider) {
		return TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBGuard, "receipt.build", "provider %s is not an available orchestration candidate for task %s", provider, taskRecord.ID)
	}
	if receipt.RecommendedDecisionLLM != "" && decisionLLM != receipt.RecommendedDecisionLLM {
		return TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBGuard, "receipt.build", "decision LLM %s does not match recommended decision LLM %s for task %s", decisionLLM, receipt.RecommendedDecisionLLM, taskRecord.ID)
	}
	receipt.GuardDecision = "accepted"
	receipt.GuardReason = "approval_id and orchestration provider candidate accepted"
	receipt.Result = "accepted"
	return receipt, nil
}
