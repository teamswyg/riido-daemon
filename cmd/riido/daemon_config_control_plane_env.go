package main

import (
	"path/filepath"
	"strings"
)

type daemonControlPlaneEnv struct {
	taskQueueDir     string
	taskReportDir    string
	taskDBSourcePath string
	saaSURL          string
	deviceID         string
	deviceSecret     string
}

func loadDaemonControlPlaneEnv(getenv func(string) string) (daemonControlPlaneEnv, error) {
	out := daemonControlPlaneEnv{
		taskQueueDir:     strings.TrimSpace(getenv(envTaskQueueDir)),
		taskReportDir:    strings.TrimSpace(getenv(envTaskReportDir)),
		taskDBSourcePath: strings.TrimSpace(getenv(envTaskDBSourcePath)),
		saaSURL:          strings.TrimSpace(getenv(envSaaSURL)),
		deviceID:         strings.TrimSpace(getenv(envDeviceID)),
		deviceSecret:     strings.TrimSpace(getenv(envDeviceSecret)),
	}
	if out.taskReportDir == "" && out.taskQueueDir != "" {
		out.taskReportDir = filepath.Join(out.taskQueueDir, "reports")
	}
	if err := validateDaemonControlPlaneMode(out); err != nil {
		return daemonControlPlaneEnv{}, err
	}
	return out, validateDaemonDeviceCredentials(out)
}

func validateDaemonControlPlaneMode(env daemonControlPlaneEnv) error {
	if env.taskDBSourcePath != "" && env.taskQueueDir != "" {
		return daemonErrorf(ErrDaemonConfig, "settings.validate-control-plane", "%s cannot be combined with %s", envTaskDBSourcePath, envTaskQueueDir)
	}
	if env.taskDBSourcePath != "" && env.taskReportDir != "" {
		return daemonErrorf(ErrDaemonConfig, "settings.validate-control-plane", "%s cannot be combined with %s", envTaskDBSourcePath, envTaskReportDir)
	}
	if env.saaSURL == "" {
		return nil
	}
	if env.taskQueueDir != "" {
		return daemonErrorf(ErrDaemonConfig, "settings.validate-control-plane", "%s cannot be combined with %s", envSaaSURL, envTaskQueueDir)
	}
	if env.taskReportDir != "" {
		return daemonErrorf(ErrDaemonConfig, "settings.validate-control-plane", "%s cannot be combined with %s", envSaaSURL, envTaskReportDir)
	}
	if env.taskDBSourcePath != "" {
		return daemonErrorf(ErrDaemonConfig, "settings.validate-control-plane", "%s cannot be combined with %s", envSaaSURL, envTaskDBSourcePath)
	}
	return nil
}

func validateDaemonDeviceCredentials(env daemonControlPlaneEnv) error {
	if env.saaSURL != "" && (env.deviceID == "" || env.deviceSecret == "") {
		return daemonErrorf(ErrDaemonConfig, "settings.validate-device-credentials", "%s requires %s/%s", envSaaSURL, envDeviceID, envDeviceSecret)
	}
	if env.deviceID == "" && env.deviceSecret != "" {
		return daemonErrorf(ErrDaemonConfig, "settings.validate-device-credentials", "%s requires %s", envDeviceSecret, envDeviceID)
	}
	if env.deviceID != "" && env.deviceSecret == "" {
		return daemonErrorf(ErrDaemonConfig, "settings.validate-device-credentials", "%s requires %s", envDeviceID, envDeviceSecret)
	}
	return nil
}
