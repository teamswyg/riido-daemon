package taskdbplane

import (
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func queuedClaimRequestDB() taskdb.TaskDB {
	return taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{queuedClaimTaskRecord()},
	}
}

func queuedClaimTaskRecord() taskdb.TaskRecord {
	return taskdb.TaskRecord{
		ID:                   "task-1",
		ProjectID:            "project-1",
		State:                task.StateQueued,
		Title:                "fallback title",
		RecommendedProvider:  "codex",
		HarnessNextDirection: "implement the patch",
		SourceDocumentPath:   "docs/task.md",
		UpdatedAt:            "2026-05-25T00:00:00Z",
	}
}
