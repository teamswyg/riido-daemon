package main

import "net/http"

func taskFixtureTeamID(client apiClient, cfg config) (string, scenario) {
	if *cfg.teamID != "" {
		return *cfg.teamID, passedTaskFixtureTeamScenario(*cfg.teamID, "configured")
	}
	path := "/workspaces/" + *cfg.workspaceID + "/teams"
	sc, payload := apiQueryPayload(client, "contract.task.fixture.team", http.MethodGet,
		path, nil, summarizeTaskFixtureTeams)
	teamID := firstTeamID(payload)
	sc.Observed["team_id"] = teamID
	if sc.Status == statusPassed && teamID == "" {
		sc.Status = statusFailed
		sc.FailureSummary = "workspace teams response did not include a team id"
		sc.Repair = taskFixtureRepair(sc.FailureSummary)
	}
	return teamID, sc
}

func passedTaskFixtureTeamScenario(teamID, source string) scenario {
	return scenario{
		ID:     "contract.task.fixture.team",
		Status: statusPassed,
		Observed: map[string]any{
			"team_id":        teamID,
			"team_id_source": source,
		},
	}
}
