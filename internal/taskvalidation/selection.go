package taskvalidation

import (
	"fmt"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func ProviderForTask(db taskdb.TaskDB, taskID, requested string) (string, error) {
	taskRecord, ok := FindTask(db, taskID)
	if !ok {
		return "", fmt.Errorf("task %s not found", taskID)
	}
	provider := strings.TrimSpace(requested)
	if provider == "" {
		provider = taskRecord.RecommendedProvider
	}
	if provider == "" {
		provider = db.RecommendedProvider
	}
	if provider == "" {
		return "", fmt.Errorf("task %s has no validation provider", taskID)
	}
	if !providerAvailable(db.ProviderCandidates, provider) {
		return "", fmt.Errorf("provider %s is not an available orchestration candidate for task %s", provider, taskID)
	}
	return provider, nil
}

func ValidateDecisionLLMForTask(db taskdb.TaskDB, taskID, requested string) error {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return nil
	}
	taskRecord, ok := FindTask(db, taskID)
	if !ok {
		return fmt.Errorf("task %s not found", taskID)
	}
	recommended := textutil.FirstNonEmpty(taskRecord.RecommendedDecisionLLM, db.RecommendedDecisionLLM)
	if recommended != "" && requested != recommended {
		return fmt.Errorf("decision LLM %s does not match recommended decision LLM %s for task %s", requested, recommended, taskID)
	}
	return nil
}

func FindTask(db taskdb.TaskDB, taskID string) (taskdb.TaskRecord, bool) {
	for _, record := range db.Tasks {
		if record.ID == taskID {
			return record, true
		}
	}
	return taskdb.TaskRecord{}, false
}

func CommandID(taskID string, now time.Time) string {
	return fmt.Sprintf("command:validation:%s:%s", taskID, now.UTC().Format("20060102T150405.000000000Z"))
}

func providerAvailable(candidates []taskdb.ProviderCandidate, provider string) bool {
	if len(candidates) == 0 {
		return true
	}
	for _, candidate := range candidates {
		if candidate.ID == provider {
			return candidate.Available
		}
	}
	return false
}
