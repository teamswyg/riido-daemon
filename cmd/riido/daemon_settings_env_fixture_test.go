package main

import "testing"

func fullDaemonSettingsEnv() map[string]string {
	return map[string]string{
		envDaemonID:                      "daemon-1",
		envDaemonVersion:                 "riido-agentd v1.2.3",
		envDaemonProfile:                 "prod",
		envServerURL:                     "https://api.riido.ai",
		envDeviceName:                    "device-a",
		envRuntimeOwner:                  "owner-a",
		envRuntimeAgents:                 "Riido, Orion, ,",
		envWorkspaceCount:                "2",
		envWorkdirRoot:                   "/tmp/riido-workspaces",
		envPolicyBundle:                  "policy-bundle.test.v1",
		envTaskQueueDir:                  "/tmp/riido-queue",
		envTaskReportDir:                 "/tmp/riido-reports",
		envWorkdirRetentionSeconds:       "86400",
		envWorkdirCleanupIntervalSeconds: "300",
		envDaemonPollIntervalSeconds:     "7",
		envDaemonIdlePollIntervalSeconds: "21",
		envDaemonHeartbeatSeconds:        "30",
	}
}

func loadDaemonSettingsForTest(t *testing.T, env map[string]string) daemonSettings {
	t.Helper()
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host-fallback", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return settings
}
