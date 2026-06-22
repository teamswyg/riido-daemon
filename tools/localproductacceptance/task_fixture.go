package main

import (
	"net/http"
	"time"
)

func maybeCreateTaskFixture(cfg config, source string) taskFixture {
	if source != "generated" || !*cfg.runMutations || !*cfg.taskFixture {
		return taskFixture{}
	}
	client := newAPIClient(*cfg.riidoAPIHost, *cfg.apiToken)
	teamID, teamScenario := taskFixtureTeamID(client, cfg)
	if teamScenario.Status != statusPassed {
		return taskFixture{Team: teamScenario}
	}
	fixture := createTaskFixture(client, teamID)
	fixture.Team = teamScenario
	return fixture
}

func createTaskFixture(client apiClient, teamID string) taskFixture {
	title := "local QA AI Agent multi-assignment " + time.Now().UTC().Format("20060102T150405Z")
	body := map[string]any{"componentType": "task", "title": title, "atOnce": true}
	sc, payload := apiQueryPayload(client, "contract.task.fixture.create", http.MethodPost,
		"/teams/"+teamID+"/components", body, summarizeTaskFixtureCreate)
	taskID := firstString(payload, "componentId", "component_id", "id")
	sc.Observed["team_id"] = teamID
	sc.Observed["task_id"] = taskID
	if sc.Status == statusPassed && taskID == "" {
		sc.Status = statusFailed
		sc.FailureSummary = "created task response did not include a task id"
		sc.Repair = taskFixtureRepair(sc.FailureSummary)
	}
	return taskFixture{TaskID: taskID, TeamID: teamID, Title: title, Create: sc}
}

func cleanupTaskFixture(cfg config, fixture taskFixture) scenario {
	client := newAPIClient(*cfg.riidoAPIHost, *cfg.apiToken)
	path := "/teams/" + fixture.TeamID + "/components/" + fixture.TaskID
	sc := apiQuery(client, "contract.task.fixture.cleanup", http.MethodDelete,
		path, nil, summarizeTaskFixtureCleanup)
	sc.Observed["team_id"] = fixture.TeamID
	sc.Observed["task_id"] = fixture.TaskID
	return sc
}
