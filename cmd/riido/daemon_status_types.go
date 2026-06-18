package main

import "github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"

// DaemonStatusSchemaVersion identifies the JSON shape returned by
// `riido daemon status` and `riido daemon health`.
const DaemonStatusSchemaVersion = "riido-agent-daemon-status.v1"

// daemonStatus is the JSON payload exposed by the daemon local socket.
type daemonStatus struct {
	SchemaVersion  string                `json:"schema_version"`
	DaemonID       string                `json:"daemon_id"`
	DaemonVersion  string                `json:"daemon_version"`
	PID            int                   `json:"pid"`
	UptimeSeconds  int                   `json:"uptime_seconds"`
	Health         string                `json:"health"`
	Ready          bool                  `json:"ready"`
	Readiness      string                `json:"readiness"`
	Profile        string                `json:"profile"`
	ServerURL      string                `json:"server_url,omitempty"`
	DeviceName     string                `json:"device_name"`
	WorkspaceCount int                   `json:"workspace_count"`
	SocketPath     string                `json:"socket_path"`
	LogFile        string                `json:"log_file,omitempty"`
	PIDFile        string                `json:"pid_file,omitempty"`
	RunningTasks   int                   `json:"running_tasks"`
	Metrics        daemonMetrics         `json:"metrics"`
	Runtimes       []runtimeactor.Status `json:"runtimes"`
	StartedAt      string                `json:"started_at"`
}

type daemonMetrics struct {
	RuntimeCount        int `json:"runtime_count"`
	RuntimeResponding   int `json:"runtime_responding"`
	ProviderAvailable   int `json:"provider_available"`
	ProviderUnavailable int `json:"provider_unavailable"`
	RunningTasks        int `json:"running_tasks"`
}
