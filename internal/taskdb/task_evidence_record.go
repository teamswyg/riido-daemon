package taskdb

import (
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func newTaskEvidenceRecord(input TaskEvidenceInput, taskRecord TaskRecord, receiptID, commandID, actor, source, stamp string, now time.Time, ordinal int) TaskEvidenceRecord {
	evidenceResult := normalizeEvidenceResult(input.Result, input.ExitCode)
	validationGate := textutil.FirstNonEmpty(input.ValidationGate, TaskEvidenceValidationV1)
	providerRunID := textutil.FirstNonEmpty(input.ProviderRunID, commandID)
	providerRunResult := textutil.FirstNonEmpty(input.ProviderRunResult, evidenceResult)
	return TaskEvidenceRecord{
		ID:                evidenceID(input.TaskID, now, ordinal),
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
		CommandReceiptID:  receiptID,
		RecordedAt:        stamp,
	}
}
