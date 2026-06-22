package main

func loginScenario(baseURL string) scenario {
	return routeScenario("product.login", baseURL, "/login")
}

func workspaceRouteScenarios(baseURL, workspaceID string) []scenario {
	if workspaceID == "" {
		return []scenario{
			configSkipped("product.runtimes.view"),
			configSkipped("product.agent.list"),
			configSkipped("product.agent.form"),
		}
	}
	prefix := "/" + workspaceID + "/settings/workspace"
	return []scenario{
		routeScenario("product.runtimes.view", baseURL, prefix+"/runtimes"),
		routeScenario("product.agent.list", baseURL, prefix+"/agents"),
		routeScenario("product.agent.form", baseURL, prefix+"/agents/new"),
	}
}

func configSkipped(id string) scenario {
	return scenario{
		ID:     id,
		Status: statusSkipped,
		Repair: &repair{
			Class:   "workspace_id_required",
			Owner:   "local-qa",
			Mode:    "manual",
			Summary: "Set RIIDO_E2E_WORKSPACE_ID to probe workspace-scoped AI Agent routes.",
		},
	}
}

func failedRoute(id, class, detail string) scenario {
	return scenario{
		ID:             id,
		Status:         statusFailed,
		FailureSummary: detail,
		Repair: &repair{
			Class:            class,
			Owner:            "frontend-runtime",
			Mode:             "manual",
			Summary:          routeRepairSummary(class),
			SuggestedCommand: routeRepairCommand(class),
		},
	}
}

func skippedRoute(id, class, detail string) scenario {
	out := failedRoute(id, class, detail)
	out.Status = statusSkipped
	return out
}
