package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

const (
	envDaemonID                      = "RIIDO_DAEMON_ID"
	envDaemonVersion                 = "RIIDO_DAEMON_VERSION"
	envDaemonProfile                 = "RIIDO_DAEMON_PROFILE"
	envServerURL                     = "RIIDO_SERVER_URL"
	envDeviceID                      = "RIIDO_DEVICE_ID"
	envDeviceSecret                  = "RIIDO_DEVICE_SECRET"
	envDeviceName                    = "RIIDO_DEVICE_NAME"
	envRuntimeOwner                  = "RIIDO_RUNTIME_OWNER"
	envRuntimeAgents                 = "RIIDO_RUNTIME_AGENTS"
	envWorkspaceCount                = "RIIDO_WORKSPACE_COUNT"
	envRuntimeMaxConcurrent          = "RIIDO_RUNTIME_MAX_CONCURRENT"
	envWorkdirRoot                   = "RIIDO_WORKDIR_ROOT"
	envPolicyBundle                  = "RIIDO_POLICY_BUNDLE_VERSION"
	envPolicyBundlePath              = "RIIDO_POLICY_BUNDLE_PATH"
	envTaskQueueDir                  = "RIIDO_TASK_QUEUE_DIR"
	envTaskReportDir                 = "RIIDO_TASK_REPORT_DIR"
	envTaskDBSourcePath              = "RIIDO_TASK_DB_SOURCE_PATH"
	envSaaSURL                       = "RIIDO_SAAS_URL"
	envDaemonPollIntervalSeconds     = "RIIDO_DAEMON_POLL_INTERVAL_SECONDS"
	envDaemonIdlePollIntervalSeconds = "RIIDO_DAEMON_IDLE_POLL_INTERVAL_SECONDS"
	envDaemonHeartbeatSeconds        = "RIIDO_DAEMON_HEARTBEAT_INTERVAL_SECONDS"
	envWorkdirRetentionSeconds       = "RIIDO_WORKDIR_RETENTION_SECONDS"
	envWorkdirCleanupIntervalSeconds = "RIIDO_WORKDIR_CLEANUP_INTERVAL_SECONDS"
)

type daemonSettings struct {
	DaemonID             string
	DaemonVersion        string
	Profile              string
	ServerURL            string
	DeviceID             string
	DeviceSecret         string
	DeviceName           string
	RuntimeOwner         string
	WorkdirRoot          string
	PolicyBundle         string
	PolicyBundlePath     string
	PolicyBundleDoc      policy.PolicyBundle
	TaskQueueDir         string
	TaskReportDir        string
	TaskDBSourcePath     string
	SaaSURL              string
	PollEvery            time.Duration
	IdlePollEvery        time.Duration
	HeartbeatEvery       time.Duration
	WorkdirRetention     time.Duration
	WorkdirCleanupEvery  time.Duration
	WorkspaceCount       int
	RuntimeMaxConcurrent int
	RuntimeAgents        []runtimeactor.AgentStatus
}

func loadDaemonSettings() (daemonSettings, error) {
	return loadDaemonSettingsFromEnvWithHome(os.Getenv, os.Hostname, os.UserHomeDir)
}

func loadDaemonSettingsFromEnv(getenv func(string) string, hostname func() (string, error)) (daemonSettings, error) {
	return loadDaemonSettingsFromEnvWithHome(getenv, hostname, os.UserHomeDir)
}

func loadDaemonSettingsFromEnvWithHome(getenv func(string) string, hostname func() (string, error), userHome func() (string, error)) (daemonSettings, error) {
	deviceName := strings.TrimSpace(getenv(envDeviceName))
	if deviceName == "" {
		if host, err := hostname(); err == nil {
			deviceName = strings.TrimSpace(host)
		}
	}
	if deviceName == "" {
		deviceName = "localhost"
	}

	workspaceCount, err := parseOptionalNonNegativeInt(getenv(envWorkspaceCount), envWorkspaceCount)
	if err != nil {
		return daemonSettings{}, err
	}

	// Max concurrent runtime sessions per provider. Default 4 so a single agent
	// isn't limited to one task at a time; override with RIIDO_RUNTIME_MAX_CONCURRENT.
	runtimeMaxConcurrent, err := parseOptionalNonNegativeInt(getenv(envRuntimeMaxConcurrent), envRuntimeMaxConcurrent)
	if err != nil {
		return daemonSettings{}, err
	}
	if runtimeMaxConcurrent == 0 {
		runtimeMaxConcurrent = 4
	}

	owner := strings.TrimSpace(getenv(envRuntimeOwner))
	if owner == "" {
		owner = strings.TrimSpace(getenv("USER"))
	}
	if owner == "" {
		owner = "local"
	}

	workdirRoot := strings.TrimSpace(getenv(envWorkdirRoot))
	if workdirRoot == "" {
		var err error
		workdirRoot, err = defaultAgentDaemonWorkdirRoot(userHome)
		if err != nil {
			return daemonSettings{}, err
		}
	}
	taskQueueDir := strings.TrimSpace(getenv(envTaskQueueDir))
	taskReportDir := strings.TrimSpace(getenv(envTaskReportDir))
	taskDBSourcePath := strings.TrimSpace(getenv(envTaskDBSourcePath))
	policyBundleVersion := strings.TrimSpace(getenv(envPolicyBundle))
	policyBundlePath := strings.TrimSpace(getenv(envPolicyBundlePath))
	policyBundleDoc := policy.DefaultLocalPolicyBundle()
	if policyBundlePath != "" {
		bundle, err := policy.LoadPolicyBundleFile(policyBundlePath)
		if err != nil {
			return daemonSettings{}, daemonWrapf(ErrDaemonConfig, "settings.load-policy-bundle", err, "load policy bundle file")
		}
		if policyBundleVersion != "" && policyBundleVersion != bundle.Version {
			return daemonSettings{}, daemonErrorf(ErrDaemonConfig, "settings.validate-policy-bundle", "%s=%q does not match %s version %q", envPolicyBundle, policyBundleVersion, envPolicyBundlePath, bundle.Version)
		}
		policyBundleDoc = bundle
		policyBundleVersion = bundle.Version
	}
	if policyBundleVersion == "" {
		policyBundleVersion = policy.DefaultLocalPolicyBundleVersion
	} else if policyBundlePath == "" {
		policyBundleDoc.Version = policyBundleVersion
	}
	saaSURL := strings.TrimSpace(getenv(envSaaSURL))
	deviceID := strings.TrimSpace(getenv(envDeviceID))
	deviceSecret := strings.TrimSpace(getenv(envDeviceSecret))
	if taskReportDir == "" && taskQueueDir != "" {
		taskReportDir = filepath.Join(taskQueueDir, "reports")
	}
	if taskDBSourcePath != "" && taskQueueDir != "" {
		return daemonSettings{}, daemonErrorf(ErrDaemonConfig, "settings.validate-control-plane", "%s cannot be combined with %s", envTaskDBSourcePath, envTaskQueueDir)
	}
	if taskDBSourcePath != "" && taskReportDir != "" {
		return daemonSettings{}, daemonErrorf(ErrDaemonConfig, "settings.validate-control-plane", "%s cannot be combined with %s", envTaskDBSourcePath, envTaskReportDir)
	}
	if saaSURL != "" {
		if taskQueueDir != "" {
			return daemonSettings{}, daemonErrorf(ErrDaemonConfig, "settings.validate-control-plane", "%s cannot be combined with %s", envSaaSURL, envTaskQueueDir)
		}
		if taskReportDir != "" {
			return daemonSettings{}, daemonErrorf(ErrDaemonConfig, "settings.validate-control-plane", "%s cannot be combined with %s", envSaaSURL, envTaskReportDir)
		}
		if taskDBSourcePath != "" {
			return daemonSettings{}, daemonErrorf(ErrDaemonConfig, "settings.validate-control-plane", "%s cannot be combined with %s", envSaaSURL, envTaskDBSourcePath)
		}
		if deviceID == "" || deviceSecret == "" {
			return daemonSettings{}, daemonErrorf(ErrDaemonConfig, "settings.validate-device-credentials", "%s requires %s/%s", envSaaSURL, envDeviceID, envDeviceSecret)
		}
	}
	if deviceID == "" && deviceSecret != "" {
		return daemonSettings{}, daemonErrorf(ErrDaemonConfig, "settings.validate-device-credentials", "%s requires %s", envDeviceSecret, envDeviceID)
	}
	if deviceID != "" && deviceSecret == "" {
		return daemonSettings{}, daemonErrorf(ErrDaemonConfig, "settings.validate-device-credentials", "%s requires %s", envDeviceID, envDeviceSecret)
	}
	workdirRetention, err := parseOptionalDurationSeconds(getenv(envWorkdirRetentionSeconds), envWorkdirRetentionSeconds)
	if err != nil {
		return daemonSettings{}, err
	}
	workdirCleanupEvery, err := parseOptionalDurationSeconds(getenv(envWorkdirCleanupIntervalSeconds), envWorkdirCleanupIntervalSeconds)
	if err != nil {
		return daemonSettings{}, err
	}
	if workdirCleanupEvery > 0 && workdirRetention <= 0 {
		return daemonSettings{}, daemonErrorf(ErrDaemonConfig, "settings.validate-workdir-retention", "%s requires %s", envWorkdirCleanupIntervalSeconds, envWorkdirRetentionSeconds)
	}
	if workdirRetention > 0 && workdirCleanupEvery <= 0 {
		workdirCleanupEvery = time.Hour
	}
	pollEvery, err := parseOptionalPositiveDurationSeconds(getenv(envDaemonPollIntervalSeconds), envDaemonPollIntervalSeconds, time.Second)
	if err != nil {
		return daemonSettings{}, err
	}
	idlePollEvery, err := parseOptionalPositiveDurationSeconds(getenv(envDaemonIdlePollIntervalSeconds), envDaemonIdlePollIntervalSeconds, 5*time.Second)
	if err != nil {
		return daemonSettings{}, err
	}
	if idlePollEvery < pollEvery {
		return daemonSettings{}, daemonErrorf(ErrDaemonConfig, "settings.validate-intervals", "%s must be greater than or equal to %s", envDaemonIdlePollIntervalSeconds, envDaemonPollIntervalSeconds)
	}
	heartbeatEvery, err := parseOptionalPositiveDurationSeconds(getenv(envDaemonHeartbeatSeconds), envDaemonHeartbeatSeconds, 5*time.Second)
	if err != nil {
		return daemonSettings{}, err
	}

	return daemonSettings{
		DaemonID:             defaultDaemonID(getenv(envDaemonID), deviceID),
		DaemonVersion:        textutil.Default(getenv(envDaemonVersion), "riido-agentd v0.0.0"),
		Profile:              textutil.Default(getenv(envDaemonProfile), "local"),
		ServerURL:            strings.TrimSpace(getenv(envServerURL)),
		DeviceID:             deviceID,
		DeviceSecret:         deviceSecret,
		DeviceName:           deviceName,
		RuntimeOwner:         owner,
		WorkdirRoot:          workdirRoot,
		PolicyBundle:         policyBundleVersion,
		PolicyBundlePath:     policyBundlePath,
		PolicyBundleDoc:      policyBundleDoc,
		TaskQueueDir:         taskQueueDir,
		TaskReportDir:        taskReportDir,
		TaskDBSourcePath:     taskDBSourcePath,
		SaaSURL:              saaSURL,
		PollEvery:            pollEvery,
		IdlePollEvery:        idlePollEvery,
		HeartbeatEvery:       heartbeatEvery,
		WorkdirRetention:     workdirRetention,
		WorkdirCleanupEvery:  workdirCleanupEvery,
		WorkspaceCount:       workspaceCount,
		RuntimeMaxConcurrent: runtimeMaxConcurrent,
		RuntimeAgents:        parseRuntimeAgents(getenv(envRuntimeAgents)),
	}, nil
}

func defaultAgentDaemonWorkdirRoot(userHome func() (string, error)) (string, error) {
	home, err := userHome()
	if err != nil {
		return "", daemonWrapf(ErrDaemonIO, "settings.default-workdir.user-home", err, "resolve user home")
	}
	root, err := hostintegration.DefaultAppDataRoot(hostintegration.AppDataRootInput{
		Channel:  hostintegration.DistributionChannelDevLocal,
		HostOS:   hostintegration.HostOSDarwin,
		UserHome: home,
	})
	if err != nil {
		return "", daemonWrapf(ErrDaemonConfig, "settings.default-workdir.app-data-root", err, "resolve default app data root")
	}
	return root.WorkdirRoot(), nil
}

func defaultDaemonID(configuredDaemonID string, deviceID string) string {
	if daemonID := strings.TrimSpace(configuredDaemonID); daemonID != "" {
		return daemonID
	}
	if devicePrincipalID := strings.TrimSpace(deviceID); devicePrincipalID != "" {
		return devicePrincipalID
	}
	return "agentd-local"
}

func parseOptionalNonNegativeInt(raw, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, daemonWrapf(ErrDaemonConfig, "settings.parse-non-negative-int", err, "%s must be a non-negative integer", name)
	}
	return n, nil
}

func parseOptionalDurationSeconds(raw, name string) (time.Duration, error) {
	n, err := parseOptionalNonNegativeInt(raw, name)
	if err != nil {
		return 0, err
	}
	return time.Duration(n) * time.Second, nil
}

func parseOptionalPositiveDurationSeconds(raw, name string, fallback time.Duration) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, daemonWrapf(ErrDaemonConfig, "settings.parse-positive-int", err, "%s must be a positive integer", name)
	}
	return time.Duration(n) * time.Second, nil
}

func parseRuntimeAgents(raw string) []runtimeactor.AgentStatus {
	parts := strings.Split(raw, ",")
	out := make([]runtimeactor.AgentStatus, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		out = append(out, runtimeactor.AgentStatus{
			AgentID: slugAgentName(name),
			Name:    name,
			State:   "online",
		})
	}
	return out
}

func slugAgentName(name string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
