package main

import "net/http"

type taskFallback struct {
	TaskID   string
	Scenario scenario
}

func existingTaskFallback(client apiClient, teamID string) taskFallback {
	if teamID == "" {
		return taskFallback{}
	}
	var last taskFallback
	for _, path := range taskFallbackPaths(teamID) {
		current := existingTaskFallbackPath(client, teamID, path)
		if current.TaskID != "" {
			return current
		}
		last = current
	}
	return last
}

func existingTaskFallbackPath(client apiClient, teamID, path string) taskFallback {
	sc, payload := apiQueryPayload(client, "contract.task.fixture.fallback_existing", http.MethodGet,
		path, nil, summarizeTaskFallback)
	taskID := firstTaskID(payload)
	sc.Observed["team_id"] = teamID
	sc.Observed["task_id"] = taskID
	sc.Observed["task_id_source"] = "readable-task-fallback"
	if sc.Status == statusPassed && taskID == "" {
		sc.Status = statusSkipped
		sc.FailureSummary = "no readable task found for fallback"
		sc.Repair = taskFixtureRepair(sc.FailureSummary)
	}
	return taskFallback{TaskID: taskID, Scenario: sc}
}

func taskFallbackPaths(teamID string) []string {
	prefix := "/teams/" + teamID + "/components"
	return []string{prefix + "/lists", prefix + "/boards", prefix + "/me"}
}

func markFixtureFallback(rows []scenario, fallback taskFallback) []scenario {
	out := append([]scenario(nil), rows...)
	for idx := range out {
		if out[idx].ID != "contract.task.fixture.create" || out[idx].Status == statusPassed {
			continue
		}
		if out[idx].Observed == nil {
			out[idx].Observed = map[string]any{}
		}
		out[idx].Status = statusSkipped
		out[idx].FailureSummary = "task fixture create unavailable; using readable task fallback"
		out[idx].Observed["fallback_task_id"] = fallback.TaskID
		out[idx].Observed["fallback_scenario_id"] = fallback.Scenario.ID
	}
	return out
}

func summarizeTaskFallback(payload map[string]any) map[string]any {
	return map[string]any{
		"task_id_present": firstTaskID(payload) != "",
	}
}
