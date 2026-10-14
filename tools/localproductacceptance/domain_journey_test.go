package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDomainJourneyScenariosUseCacheAndRelatedEvidence(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "domain-cache.json")
	body := `{"schema_version":"riido-domain-fixture-cache.v1","id":"domain-fixture-journey","environment":"staging","entities":{"project":{"id":"project-a"}}}`
	if err := os.WriteFile(cachePath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := domainJourneyTestConfig(cachePath)
	got := domainJourneyScenarios(cfg, []scenario{{ID: "contract.task.thread_message", Status: statusPassed}})
	if got[0].ID != "domain.fixture_journey" || got[0].Observed["common_name"] != "도메인 픽스처 여정" {
		t.Fatalf("summary=%+v", got[0])
	}
	if findDomainScenario(got, "domain.fixture.project").Status != statusPassed {
		t.Fatalf("project row=%+v", findDomainScenario(got, "domain.fixture.project"))
	}
	if findDomainScenario(got, "domain.fixture.thread").Status != statusPassed {
		t.Fatalf("thread row=%+v", findDomainScenario(got, "domain.fixture.thread"))
	}
}

func TestDomainJourneyScenariosRequireMissingFixtureEvidence(t *testing.T) {
	cfg := domainJourneyTestConfig(filepath.Join(t.TempDir(), "missing.json"))
	got := domainJourneyScenarios(cfg, nil)
	if findDomainScenario(got, "domain.fixture.milestone").Status != statusSkipped {
		t.Fatalf("milestone row=%+v", findDomainScenario(got, "domain.fixture.milestone"))
	}
	if findDomainScenario(got, "domain.fixture.milestone").Repair == nil {
		t.Fatalf("repair missing: %+v", findDomainScenario(got, "domain.fixture.milestone"))
	}
}

func domainJourneyTestConfig(cache string) config {
	base, token, workspace, task := "http://127.0.0.1:3000", "token", "workspace-a", "task-a"
	first, second, team := "agent-a", "", "team-a"
	agentHost, riidoHost := "https://staging.ai-api.riido.io", "https://staging.api.riido.io"
	d := 24 * time.Hour
	return config{
		baseURL:       &base,
		apiToken:      &token,
		workspaceID:   &workspace,
		taskID:        &task,
		firstAgentID:  &first,
		secondAgentID: &second,
		teamID:        &team,
		agentHost:     &agentHost,
		riidoAPIHost:  &riidoHost,
		domainCache:   &cache,
		validFor:      &d,
	}
}

func findDomainScenario(rows []scenario, id string) scenario {
	for _, row := range rows {
		if row.ID == id {
			return row
		}
	}
	return scenario{}
}
