package main

import "testing"

func TestAPIConfigAndSkippedScenarios(t *testing.T) {
	empty := ""
	token := "tok"
	host := "https://staging.ai-api.riido.io"
	cfg := config{workspaceID: &empty, apiToken: &token, agentHost: &host}
	if repair := missingAPIConfig(cfg); repair == nil || repair.Class != "workspace_id_required" {
		t.Fatalf("expected workspace repair, got %#v", repair)
	}
	ws := "workspace with/slash"
	cfg.workspaceID = &ws
	if missingAPIConfig(cfg) != nil {
		t.Fatal("complete API config should not be missing")
	}
	if got := workspaceBase(ws); got != "/v2/client/workspaces/workspace%20with%2Fslash/ai-agent" {
		t.Fatalf("unexpected workspace base: %s", got)
	}
	body := thumbnailIntentBody()
	if body["content_type"] != "image/png" || body["file_name"] == "" {
		t.Fatalf("unexpected upload intent body: %#v", body)
	}
	skipped := apiSkippedScenarios(apiRuntimeRepair())
	if len(skipped) != 8 || skipped[0].Status != statusSkipped {
		t.Fatalf("unexpected skipped scenarios: %#v", skipped)
	}
}

func TestAPIRepairsClassifyRuntimeBinding(t *testing.T) {
	repair := apiRepairForPayload(map[string]any{
		"error": "ai agent runtime binding is not configured",
	})
	if repair.Class != "ai_agent_runtime_binding_required" {
		t.Fatalf("expected binding repair, got %#v", repair)
	}
	generic := apiRepairForPayload(map[string]any{"error": "unauthorized"})
	if generic.Class != "staging_api_unavailable_or_unauthorized" {
		t.Fatalf("expected runtime repair, got %#v", generic)
	}
	sc := failTaskScenario("contract.task.thread_message", "task id missing")
	if sc.Status != statusSkipped || sc.Repair.Class != "task_flow_config_required" {
		t.Fatalf("unexpected failed task scenario: %#v", sc)
	}
}
