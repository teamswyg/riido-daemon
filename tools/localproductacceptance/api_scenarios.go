package main

import (
	"net/http"
	"net/url"
)

func contractAPIScenarios(cfg config) []scenario {
	base := workspaceBase(*cfg.workspaceID)
	if missing := missingAPIConfig(cfg); missing != nil {
		return apiSkippedScenarios(missing)
	}
	client := newAPIClient(*cfg.agentHost, *cfg.apiToken)
	prep := prepareSaaSDaemons(cfg, client)
	prep = waitForPreparedSaaSRuntimes(cfg, client, prep)
	discovery, discoveryPayload := apiQueryPayload(client, "contract.task.discovery", http.MethodGet,
		base+"/tasks/assigned-agent-profiles", nil, summarizeAssignedProfiles)
	devices, _ := apiQueryPayload(client, "contract.api.devices", http.MethodGet,
		base+"/devices", nil, summarizeDevices)
	agents := maybeCreateAgentFixtures(cfg, client, base, prep)
	out := []scenario{
		apiQuery(client, "contract.api.bootstrap", http.MethodGet, base+"/bootstrap", nil, summarizeBootstrap),
	}
	out = append(out, prep.Scenarios...)
	out = append(out, devices)
	out = append(out, agents.Scenarios...)
	out = append(out,
		apiQuery(client, "contract.api.profile_thumbnail.intent", http.MethodPost,
			base+"/profile-thumbnails/uploads", thumbnailIntentBody(), summarizeUploadIntent),
		discovery,
	)
	out = append(out, taskFlowScenarios(client, cfg, discoveryPayload, agents)...)
	out = append(out, cleanupSaaSDaemons(*cfg.daemonBinary, prep)...)
	return out
}

func missingAPIConfig(cfg config) *repair {
	if *cfg.workspaceID == "" {
		return apiConfigRepair("workspace_id_required", "Set RIIDO_E2E_WORKSPACE_ID or provide a storage-state workspace key.")
	}
	if *cfg.apiToken == "" {
		return apiConfigRepair("api_token_required", "Set RIIDO_AI_AGENT_TOKEN or provide a storage-state token cookie.")
	}
	return nil
}

func workspaceBase(workspaceID string) string {
	return "/v2/client/workspaces/" + url.PathEscape(workspaceID) + "/ai-agent"
}

func thumbnailIntentBody() map[string]any {
	return map[string]any{
		"content_type":         "image/png",
		"content_length_bytes": 128,
		"file_name":            "local-contract-lab.png",
	}
}

type summarizeFunc func(map[string]any) map[string]any
