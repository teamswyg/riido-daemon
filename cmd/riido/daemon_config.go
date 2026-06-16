package main

import (
	"os"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/policy"
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
	envDaemonPprofAddr               = "RIIDO_DAEMON_PPROF_ADDR"
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
	PprofAddr            string
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
