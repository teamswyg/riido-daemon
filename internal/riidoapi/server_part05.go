package riidoapi

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func validationProviderForTask(db taskdb.TaskDB, taskID, requested string) (string, error) {
	taskRecord, ok := findTask(db, taskID)
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

func validateDecisionLLMForTask(db taskdb.TaskDB, taskID, requested string) error {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return nil
	}
	taskRecord, ok := findTask(db, taskID)
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

func rawParams(params any) (json.RawMessage, error) {
	if params == nil {
		return nil, nil
	}
	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("encode riido API params: %w", err)
	}
	return data, nil
}

func okResponse(method string, data any) responseEnvelope {
	payload, err := json.Marshal(data)
	if err != nil {
		return errorResponse(method, err)
	}
	return responseEnvelope{
		OK:     true,
		Method: method,
		Data:   payload,
	}
}

func errorResponse(method string, err error) responseEnvelope {
	return responseEnvelope{
		OK:     false,
		Method: method,
		Error:  err.Error(),
	}
}

func writeResponse(conn net.Conn, response responseEnvelope) error {
	encoder := json.NewEncoder(conn)
	return encoder.Encode(response)
}

type requestEnvelope struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

type responseEnvelope struct {
	OK     bool            `json:"ok"`
	Method string          `json:"method"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
}
