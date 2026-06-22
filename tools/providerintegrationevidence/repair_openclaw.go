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
		Class:            "openclaw_model_config_required",
		Owner:            "local_operator",
		Mode:             "manual",
		Summary:          "OpenClaw ran but did not perform the required file side effect; point OpenClaw at a fast tool-capable local model and rerun QA.",
		SuggestedCommand: "openclaw doctor --fix && openclaw models set llama3.2:latest && go run ./tools/localqarunner -run-product",
	}
}
