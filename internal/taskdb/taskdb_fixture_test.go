package taskdb

import (
	"time"

	"github.com/teamswyg/riido-contracts/task"
)

var testMinute = time.Minute

func sampleTaskDB() TaskDB {
	return TaskDB{
		SchemaVersion:          TaskDBSchemaVersion,
		DecisionGate:           "human-approval-required",
		RecommendedProvider:    "codex",
		RecommendedDecisionLLM: "codex",
		ProviderCandidates: []ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []TaskRecord{sampleTaskRecord()},
	}
}

func sampleTaskRecord() TaskRecord {
	return TaskRecord{
		ID:                     "task-1",
		ProjectID:              "project-1",
		State:                  task.StateCreated,
		Title:                  "Implement task",
		RecommendedProvider:    "codex",
		RecommendedDecisionLLM: "codex",
		RequiresHumanApproval:  true,
		SourceDocumentID:       "doc-1",
		SourceDocumentPath:     "docs/task.md",
	}
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 28, 1, 2, 3, 0, time.UTC)
}
