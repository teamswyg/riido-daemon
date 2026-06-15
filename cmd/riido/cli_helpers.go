package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func validationProviderForTask(db taskdb.TaskDB, taskID, requested string) (string, error) {
	taskRecord, ok := findTaskRecord(db, taskID)
	if !ok {
		return "", fmt.Errorf("task %s not found", taskID)
	}
	provider := requested
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

func validateDecisionLLMForTask(db taskdb.TaskDB, taskID, requested string) error {
	if requested == "" {
		return nil
	}
	taskRecord, ok := findTaskRecord(db, taskID)
	if !ok {
		return fmt.Errorf("task %s not found", taskID)
	}
	recommended := taskRecord.RecommendedDecisionLLM
	if recommended == "" {
		recommended = db.RecommendedDecisionLLM
	}
	if recommended != "" && requested != recommended {
		return fmt.Errorf("decision LLM %s does not match recommended decision LLM %s for task %s", requested, recommended, taskID)
	}
	return nil
}

func findTaskRecord(db taskdb.TaskDB, taskID string) (taskdb.TaskRecord, bool) {
	for _, record := range db.Tasks {
		if record.ID == taskID {
			return record, true
		}
	}
	return taskdb.TaskRecord{}, false
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

func validationCommandID(taskID string, now time.Time) string {
	return fmt.Sprintf("command:validation:%s:%s", taskID, now.UTC().Format("20060102T150405.000000000Z"))
}

func validationTransitionForResult(result string) (task.TaskState, ir.EventType) {
	if result == "passed" {
		return task.StatePatchReady, ir.EventValidationPassed
	}
	return task.StateFailed, ir.EventValidationFailed
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  riido serve [--socket PATH] [--transport unix-socket|windows-named-pipe] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido api <status|tasks> [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido api review-demo --channel CHANNEL --review-demo-consent-granted true|false [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido api transition <task-id> --to STATE --event EVENT --approval-id ID [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--actor ACTOR] [--source SOURCE] [--reason TEXT] [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido api evidence <task-id> --command COMMAND --approval-id ID [--exit-code N] [--result RESULT] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--provider-run-id ID] [--provider-run-result RESULT] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido api validate <task-id> --command COMMAND --approval-id ID [--workdir PATH] [--timeout-seconds N] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido mwsd <snapshot|projection|sync|orchestration|projects|status> [--socket PATH] [--state PATH] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido task list [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido task transition <task-id> --to STATE --event EVENT --approval-id ID [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--actor ACTOR] [--source SOURCE] [--reason TEXT] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido task evidence <task-id> --command COMMAND --approval-id ID [--exit-code N] [--result RESULT] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--provider-run-id ID] [--provider-run-result RESULT] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido task validate <task-id> --command COMMAND --approval-id ID [--workdir PATH] [--timeout-seconds N] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido bridge <providers|detect>")
	fmt.Fprintln(os.Stderr, "  riido daemon start [--foreground] [--socket PATH] [--pid-file PATH] [--log-file PATH] [--lock-file PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon status [--socket PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon health [--socket PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon ready [--socket PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon metrics [--socket PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon stop [--socket PATH] [--pid-file PATH] [--timeout-seconds N] [--force]")
	fmt.Fprintln(os.Stderr, "  riido daemon logs --log-file PATH [--lines N]")
}
