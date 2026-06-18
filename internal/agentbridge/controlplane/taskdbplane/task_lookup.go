package taskdbplane

import (
	"slices"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func findTask(db taskdb.TaskDB, taskID string) (taskdb.TaskRecord, bool) {
	for _, record := range db.Tasks {
		if record.ID == taskID {
			return record, true
		}
	}
	return taskdb.TaskRecord{}, false
}

func providerFor(db taskdb.TaskDB, record taskdb.TaskRecord) string {
	return textutil.FirstNonEmptyTrimmed(record.RecommendedProvider, db.RecommendedProvider)
}

func decisionLLMFor(db taskdb.TaskDB, record taskdb.TaskRecord) string {
	return textutil.FirstNonEmptyTrimmed(record.RecommendedDecisionLLM, db.RecommendedDecisionLLM)
}

func promptFor(record taskdb.TaskRecord) string {
	return textutil.FirstNonEmptyTrimmed(record.HarnessNextDirection, record.Title)
}

func requiresApproval(db taskdb.TaskDB, record taskdb.TaskRecord) bool {
	return record.RequiresHumanApproval || db.DecisionGate == "human-approval-required"
}

func approvalIDForTask(db taskdb.TaskDB, taskID string) string {
	for _, receipt := range slices.Backward(db.CommandReceipts) {
		if receipt.TaskID == taskID && strings.TrimSpace(receipt.ApprovalID) != "" {
			return strings.TrimSpace(receipt.ApprovalID)
		}
	}
	return ""
}

func providerAvailable(db taskdb.TaskDB, provider string) bool {
	if len(db.ProviderCandidates) == 0 {
		return true
	}
	for _, candidate := range db.ProviderCandidates {
		if candidate.ID == provider {
			return candidate.Available
		}
	}
	return false
}
