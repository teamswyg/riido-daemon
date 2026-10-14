package project

import "github.com/teamswyg/riido-daemon/internal/taskdb"

func applyTaskDBSource(record taskdb.TaskRecord, sourceTask TaskState, stamp string) taskdb.TaskRecord {
	record.ProjectID = sourceTask.ProjectID
	record.SourceDocumentID = sourceTask.SourceDocumentID
	record.SourceDocumentPath = sourceTask.SourceDocumentPath
	record.Title = sourceTask.Title
	record.Owner = sourceTask.Owner
	record.SourceStatus = sourceTask.SourceStatus
	record.RecommendedProvider = sourceTask.RecommendedProvider
	record.RecommendedDecisionLLM = sourceTask.RecommendedDecisionLLM
	record.RequiresHumanApproval = sourceTask.RequiresHumanApproval
	record.HarnessNextDirection = sourceTask.HarnessNextDirection
	record.UpdatedAt = stamp
	return record
}
