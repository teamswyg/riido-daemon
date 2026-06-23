package main

import (
	"encoding/json"
	"os"
	"strings"
)

const domainJourneyID = "domain-fixture-journey"

type domainFixtureCache struct {
	SchemaVersion string                        `json:"schema_version"`
	ID            string                        `json:"id"`
	Environment   string                        `json:"environment"`
	Entities      map[string]domainCachedEntity `json:"entities"`
}

type domainCachedEntity struct {
	ID     string `json:"id"`
	Name   string `json:"name,omitempty"`
	Source string `json:"source,omitempty"`
}

type domainEntityDef struct {
	Key       string
	Title     string
	Create    string
	Verify    string
	CacheKey  string
	RelatedID string
}

func domainJourneyScenarios(cfg config, prior []scenario) []scenario {
	cache := readDomainFixtureCache(*cfg.domainCache)
	entities := domainEntityDefs()
	rows := make([]scenario, 0, len(entities)+1)
	rows = append(rows, domainJourneySummary(cfg, cache, prior, entities))
	for _, entity := range entities {
		rows = append(rows, domainEntityScenario(cfg, cache, prior, entity))
	}
	return rows
}

func domainJourneySummary(cfg config, cache domainFixtureCache, prior []scenario, entities []domainEntityDef) scenario {
	remote := domainRemoteEnvironment(*cfg.riidoAPIHost, *cfg.agentHost)
	verification := domainVerificationSource(*cfg.baseURL)
	row := scenario{
		ID:     "domain.fixture_journey",
		Status: statusPassed,
		Observed: map[string]any{
			"id":                  domainJourneyID,
			"name":                "Domain Fixture Journey",
			"common_name":         "도메인 픽스처 여정",
			"remote_environment":  remote,
			"remote_policy":       "staging-only",
			"verification_source": verification,
			"cache_path":          *cfg.domainCache,
			"cache_loaded":        cache.ID != "",
			"meaningful_entities": domainEntityKeys(entities),
			"passed_related_rows": passedScenarioIDs(prior),
			"daily_minimum":       "covered by local QA daily schedule; stale after expires_at",
			"extension_rule":      "add a domainEntityDef and a DSL entity row for each new product object",
		},
	}
	if remote != "staging" || verification != "local" {
		row.Status = statusSkipped
		row.FailureSummary = "domain fixture journey requires staging remote hosts and local verification"
		row.Repair = &repair{
			Class:   "domain_fixture_environment_required",
			Owner:   "local-qa",
			Mode:    "configuration",
			Summary: "Run with staging AI/API hosts while the frontend and daemon verification surface runs locally.",
		}
	}
	return row
}

func domainEntityScenario(cfg config, cache domainFixtureCache, prior []scenario, entity domainEntityDef) scenario {
	cached := cache.Entities[entity.Key]
	status := statusSkipped
	source := "missing"
	switch {
	case cached.ID != "":
		status, source = statusPassed, "cache"
	case configuredDomainID(cfg, entity.Key) != "":
		status, source = statusPassed, "configured"
	case relatedScenarioPassed(prior, entity.RelatedID):
		status, source = statusPassed, "contract-evidence"
	}
	row := scenario{
		ID:       "domain.fixture." + entity.Key,
		Status:   status,
		Method:   "DOMAIN",
		Endpoint: entity.Verify,
		Observed: map[string]any{
			"title":              entity.Title,
			"cache_key":          entity.CacheKey,
			"cached_id":          cached.ID,
			"source":             source,
			"create_endpoint":    entity.Create,
			"verify_endpoint":    entity.Verify,
			"remote_environment": domainRemoteEnvironment(*cfg.riidoAPIHost, *cfg.agentHost),
			"cache_path":         *cfg.domainCache,
			"related_scenario":   entity.RelatedID,
		},
	}
	if status != statusPassed {
		row.FailureSummary = "no cached or observed staging fixture id yet"
		row.Repair = &repair{
			Class:   "domain_fixture_required",
			Owner:   "local-qa",
			Mode:    "manual-or-automated",
			Summary: "Create or import the staging " + entity.Title + " fixture, then export domain-fixture-journey-cache.json.",
		}
	}
	return row
}

func domainEntityDefs() []domainEntityDef {
	return []domainEntityDef{
		{"account", "Account", "POST /users (guarded signup fixture)", "GET /users/me", "account_id", ""},
		{"workspace", "Workspace", "POST /workspaces", "GET /workspaces/{workspace_id}", "workspace_id", "contract.api.bootstrap"},
		{"team", "Team", "POST /workspaces/{workspace_id}/teams", "GET /workspaces/{workspace_id}/teams/{team_key}", "team_id", "contract.task.fixture.team"},
		{"project", "Project", "POST /teams/{team_id}/components componentType=project", "GET /teams/{team_id}/components/lists", "project_id", ""},
		{"milestone", "Milestone", "POST /teams/{team_id}/components componentType=milestone", "GET /teams/{team_id}/components/lists", "milestone_id", ""},
		{"task", "Task", "POST /teams/{team_id}/components componentType=task", "GET /tasks/{task_id}/assignable-agents", "task_id", "contract.task.fixture.create"},
		{"agent", "Agent", "POST /v2/client/workspaces/{workspace_id}/ai-agent/agents", "GET /v2/client/workspaces/{workspace_id}/ai-agent/devices", "agent_id", "local.saas.agent_fixture.create.1"},
		{"thread", "Thread", "POST /tasks/{task_id}/agent-assignments", "GET /tasks/{task_id}/thread-stream-subscription", "thread_id", "contract.task.thread_message"},
	}
}

func readDomainFixtureCache(path string) domainFixtureCache {
	data, err := os.ReadFile(path)
	if err != nil {
		return domainFixtureCache{}
	}
	var cache domainFixtureCache
	if json.Unmarshal(data, &cache) != nil || cache.Entities == nil {
		return domainFixtureCache{}
	}
	return cache
}

func configuredDomainID(cfg config, key string) string {
	switch key {
	case "workspace":
		return strings.TrimSpace(*cfg.workspaceID)
	case "team":
		return strings.TrimSpace(*cfg.teamID)
	case "task":
		return strings.TrimSpace(*cfg.taskID)
	case "agent":
		if strings.TrimSpace(*cfg.firstAgentID) != "" {
			return strings.TrimSpace(*cfg.firstAgentID)
		}
		return strings.TrimSpace(*cfg.secondAgentID)
	default:
		return ""
	}
}

func relatedScenarioPassed(rows []scenario, id string) bool {
	if id == "" {
		return false
	}
	for _, row := range rows {
		if row.ID == id && row.Status == statusPassed {
			return true
		}
	}
	return false
}

func passedScenarioIDs(rows []scenario) []string {
	out := []string{}
	for _, row := range rows {
		if row.Status == statusPassed {
			out = append(out, row.ID)
		}
	}
	return out
}

func domainEntityKeys(entities []domainEntityDef) []string {
	out := make([]string, 0, len(entities))
	for _, entity := range entities {
		out = append(out, entity.Key)
	}
	return out
}

func domainRemoteEnvironment(riidoHost, agentHost string) string {
	joined := strings.ToLower(riidoHost + " " + agentHost)
	switch {
	case strings.Contains(joined, "staging"):
		return "staging"
	case strings.Contains(joined, "production"):
		return "production"
	case strings.Contains(joined, "development"):
		return "development"
	default:
		return "custom"
	}
}

func domainVerificationSource(baseURL string) string {
	base := strings.ToLower(baseURL)
	if strings.Contains(base, "localhost") || strings.Contains(base, "127.0.0.1") {
		return "local"
	}
	return "custom"
}
