package main

import "testing"

func TestLoadDaemonSettingsReadsIdentity(t *testing.T) {
	settings := loadDaemonSettingsForTest(t, fullDaemonSettingsEnv())
	if settings.DaemonID != "daemon-1" ||
		settings.DaemonVersion != "riido-agentd v1.2.3" ||
		settings.Profile != "prod" ||
		settings.ServerURL != "https://api.riido.ai" {
		t.Fatalf("daemon fields: %+v", settings)
	}
	if settings.DeviceName != "device-a" || settings.RuntimeOwner != "owner-a" || settings.WorkspaceCount != 2 {
		t.Fatalf("settings mismatch: %+v", settings)
	}
	if len(settings.RuntimeAgents) != 2 ||
		settings.RuntimeAgents[0].AgentID != "riido" ||
		settings.RuntimeAgents[1].AgentID != "orion" {
		t.Fatalf("agents: %+v", settings.RuntimeAgents)
	}
}
