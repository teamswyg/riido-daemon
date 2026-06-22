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
