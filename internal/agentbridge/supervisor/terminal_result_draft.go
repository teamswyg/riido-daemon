package supervisor

import (
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func terminalResultDraft(res agentbridge.Result) (ir.EventType, map[string]any) {
	status := res.Status
	if status == "" {
		status = agentbridge.ResultCompleted
	}
	switch status {
	case agentbridge.ResultCompleted:
		return ir.EventRunReportedDone, map[string]any{
			"summary":      res.Output,
			"resultStatus": string(status),
		}
	case agentbridge.ResultCancelled:
		return ir.EventTaskCancelled, map[string]any{
			"reason":  textutil.FirstNonEmpty(res.Error, "provider run cancelled"),
			"byActor": "daemon",
		}
	case agentbridge.ResultTimeout:
		return ir.EventTaskTimedOut, timeoutPayload(res)
	default:
		return ir.EventTaskFailed, map[string]any{
			"category": taskFailureCategory(status),
			"reason":   textutil.FirstNonEmpty(res.Error, string(status)),
			"terminal": true,
		}
	}
}

func timeoutPayload(res agentbridge.Result) map[string]any {
	payload := map[string]any{
		"fromState": "Running",
		"limit":     textutil.FirstNonEmpty(res.Error, "timeout"),
	}
	if !res.StartedAt.IsZero() && !res.FinishedAt.IsZero() {
		payload["elapsed"] = res.FinishedAt.Sub(res.StartedAt).String()
	}
	return payload
}

func taskFailureCategory(status agentbridge.ResultStatus) string {
	switch status {
	case agentbridge.ResultBlocked:
		return "provider_blocked"
	case agentbridge.ResultAborted:
		return "process_aborted"
	default:
		return "provider_result_failed"
	}
}
