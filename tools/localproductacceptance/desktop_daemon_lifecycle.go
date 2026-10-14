package main

import "os"

func desktopDaemonLifecycleScenario() scenario {
	stopPath := desktopDaemonStopEvidencePath()
	socketPath := desktopAgentdSocketPath()
	events, err := readDesktopDaemonStopEvents(stopPath)
	observed := map[string]any{
		"stop_evidence_path": stopPath,
		"agentd_socket_path": socketPath,
		"agentd_socket_live": localFileExists(socketPath),
	}
	if err != nil {
		if os.IsNotExist(err) {
			observed["stop_evidence_present"] = false
			return scenario{ID: "local.daemon.desktop_shutdown_lifecycle", Status: statusPassed, Observed: observed}
		}
		return failTaskScenario("local.daemon.desktop_shutdown_lifecycle", "read desktop daemon stop evidence: "+err.Error())
	}
	observed["stop_evidence_present"] = true
	observed["stop_event_count"] = len(events)
	if len(events) == 0 {
		return scenario{ID: "local.daemon.desktop_shutdown_lifecycle", Status: statusPassed, Observed: observed}
	}
	last := events[len(events)-1]
	observed["last_stop_reason"] = last.Reason
	observed["last_stop_method"] = last.Method
	observed["last_stop_profile"] = last.Profile
	observed["last_stop_observed_at"] = last.ObservedAt
	if last.Reason == "app-quit" && !localFileExists(socketPath) {
		return desktopDaemonShutdownCandidate(observed)
	}
	return scenario{ID: "local.daemon.desktop_shutdown_lifecycle", Status: statusPassed, Observed: observed}
}

func desktopDaemonShutdownCandidate(observed map[string]any) scenario {
	return scenario{
		ID:       "local.daemon.desktop_shutdown_lifecycle",
		Status:   statusPartial,
		Observed: observed,
		Repair: &repair{
			Class:            "daemon_lifecycle_policy",
			Owner:            "desktop/daemon",
			Mode:             "closed_loop_candidate",
			Summary:          "Desktop app-quit stopped the AI Agent daemon and no agentd socket is currently live.",
			SuggestedCommand: "Bind desktop app-quit daemon shutdown to a business claim, verifier, and relaunch/adopt decision.",
		},
	}
}
