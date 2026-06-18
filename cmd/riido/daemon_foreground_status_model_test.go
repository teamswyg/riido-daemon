package main

type daemonStatusJSON struct {
	SchemaVersion  string                `json:"schema_version"`
	DaemonID       string                `json:"daemon_id"`
	DaemonVersion  string                `json:"daemon_version"`
	PID            int                   `json:"pid"`
	UptimeSeconds  int                   `json:"uptime_seconds"`
	Health         string                `json:"health"`
	Ready          bool                  `json:"ready"`
	Readiness      string                `json:"readiness"`
	Profile        string                `json:"profile"`
	ServerURL      string                `json:"server_url"`
	DeviceName     string                `json:"device_name"`
	WorkspaceCount int                   `json:"workspace_count"`
	SocketPath     string                `json:"socket_path"`
	RunningTasks   int                   `json:"running_tasks"`
	Metrics        daemonStatusMetrics   `json:"metrics"`
	Runtimes       []daemonRuntimeStatus `json:"runtimes"`
}

type daemonStatusMetrics struct {
	RuntimeCount        int `json:"runtime_count"`
	RuntimeResponding   int `json:"runtime_responding"`
	ProviderAvailable   int `json:"provider_available"`
	ProviderUnavailable int `json:"provider_unavailable"`
	RunningTasks        int `json:"running_tasks"`
}

type daemonRuntimeStatus struct {
	RuntimeID     string                   `json:"runtime_id"`
	Health        string                   `json:"health"`
	Owner         string                   `json:"owner"`
	DeviceName    string                   `json:"device_name"`
	Agents        []daemonAgentStatus      `json:"agents"`
	Capabilities  []daemonCapabilityStatus `json:"capabilities"`
	MaxConcurrent int                      `json:"max_concurrent"`
}

type daemonAgentStatus struct {
	AgentID string `json:"agent_id"`
	Name    string `json:"name"`
	State   string `json:"state"`
}

type daemonCapabilityStatus struct {
	Provider              string `json:"provider"`
	Available             bool   `json:"available"`
	Reason                string `json:"reason"`
	ProtocolKind          string `json:"protocol_kind"`
	AdapterID             string `json:"adapter_id"`
	AdapterVersion        string `json:"adapter_version"`
	ProtocolVersion       string `json:"protocol_version"`
	CompatibilityStatus   string `json:"compatibility_status"`
	CapabilityFingerprint string `json:"capability_fingerprint"`
}
