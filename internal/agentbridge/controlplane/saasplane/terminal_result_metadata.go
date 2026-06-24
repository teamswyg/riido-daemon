package saasplane

import (
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func terminalResultMetadata(res agentbridge.Result, message string) map[string]string {
	status := res.Status
	if status == "" {
		status = agentbridge.ResultCompleted
	}
	if terminalResultNeedsUserInput(res, message) {
		status = "needs_input"
	}
	metadata := map[string]string{
		metadatakeys.AssignmentResultStatus.String(): string(status),
	}
	if category := terminalFailureCategory(status); category != "" {
		metadata[metadatakeys.AssignmentFailureCategory.String()] = category
	}
	return metadata
}

func terminalFailureCategory(status agentbridge.ResultStatus) string {
	switch status {
	case agentbridge.ResultBlocked:
		return "provider_blocked"
	case agentbridge.ResultAborted:
		return "process_aborted"
	case agentbridge.ResultTimeout:
		return "provider_timeout"
	case agentbridge.ResultFailed:
		return "provider_result_failed"
	default:
		return ""
	}
}

func terminalResultNeedsUserInput(res agentbridge.Result, message string) bool {
	status := res.Status
	if status == "" {
		status = agentbridge.ResultCompleted
	}
	if status != agentbridge.ResultCompleted {
		return false
	}
	normalized := strings.ToLower(strings.TrimSpace(message))
	if normalized == "" {
		normalized = strings.ToLower(strings.TrimSpace(res.Output))
	}
	for _, marker := range needsUserInputMarkers() {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}
