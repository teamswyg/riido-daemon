package main

func openClawConfigRepair() repair {
	return repair{
		Class:            "provider_config_invalid",
		Owner:            "human",
		Mode:             "manual",
		Summary:          "OpenClaw config is invalid for the installed CLI version.",
		SuggestedCommand: "openclaw doctor --fix",
	}
}

func openClawBackendRepair() repair {
	return repair{
		Class:            "local_backend_unavailable",
		Owner:            "local_operator",
		Mode:             "candidate_auto",
		Summary:          "OpenClaw local model backend is unavailable or unhealthy.",
		SuggestedCommand: "brew services start ollama || ollama serve",
	}
}

func openClawModelConfigRepair() repair {
	return repair{
		Class:            "openclaw_cwd_side_effect_unverified",
		Owner:            "local_operator",
		Mode:             "manual",
		Summary:          "OpenClaw ran, but filesystem side effects in the daemon-selected cwd are not verified for this local configuration.",
		SuggestedCommand: "openclaw doctor --fix && openclaw models set ollama/llama3.2:latest && AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/openclaw -race -count=1 -run TestIntegration -v",
	}
}
