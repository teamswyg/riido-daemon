package main

import "strings"

func sseReplayScenario(client apiClient, base string, first, second scenario) scenario {
	out := scenario{ID: "contract.task.sse_replay", Method: "GET", Endpoint: base + "/events?replay=1"}
	text, status, err := apiStreamReplay(client, out.Endpoint)
	out.Observed = map[string]any{"status_code": status, "bytes": len(text)}
	if err != nil {
		out.Status = statusFailed
		out.FailureSummary = err.Error()
		out.Repair = apiRuntimeRepair()
		return out
	}
	out.Observed["first_assignment_seen"] = replayContainsScenario(text, first)
	out.Observed["second_assignment_seen"] = replayContainsScenario(text, second)
	if out.Observed["first_assignment_seen"] != true || out.Observed["second_assignment_seen"] != true {
		out.Status = statusFailed
		out.FailureSummary = "SSE replay did not include both assignment identifiers"
		return out
	}
	out.Status = statusPassed
	return out
}

func replayContainsScenario(text string, scenario scenario) bool {
	assignmentID, _ := scenario.Observed["assignment_id"].(string)
	threadID, _ := scenario.Observed["thread_id"].(string)
	return assignmentID != "" && strings.Contains(text, assignmentID) ||
		threadID != "" && strings.Contains(text, threadID)
}
