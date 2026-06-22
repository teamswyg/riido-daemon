package main

func routeRepairSummary(class string) string {
	switch class {
	case "frontend_not_running":
		return "Start the local frontend as an external target; do not commit harness code to riido-client."
	case "frontend_route_missing":
		return "The local frontend does not expose the expected AI Agent product route."
	case "frontend_auth_required":
		return "Authenticated browser evidence is required for this route."
	default:
		return "Inspect the local frontend route and product acceptance evidence."
	}
}

func routeRepairCommand(class string) string {
	if class != "frontend_not_running" {
		return ""
	}
	return "cd ../riido-client && NEXT_PUBLIC_AI_AGENT_HOST=https://development.ai-api.riido.io pnpm run dev"
}
