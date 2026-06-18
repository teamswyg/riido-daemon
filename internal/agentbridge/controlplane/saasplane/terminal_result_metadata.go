package saasplane

import (
	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func terminalResultMetadata(res agentbridge.Result) map[string]string {
	status := res.Status
	if status == "" {
		status = agentbridge.ResultCompleted
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
