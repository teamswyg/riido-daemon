package runtimeactor

import "time"

// TaskStatus describes one in-flight task within the runtime.
type TaskStatus struct {
	TaskID    string `json:"task_id"`
	Provider  string `json:"provider"`
	SessionID string `json:"session_id,omitempty"`
	State     string `json:"state"`
}

// AgentStatus is the runtime -> agent association shown by the local
// settings UI. The runtime actor does not schedule per-agent work yet;
// it simply publishes the binding data supplied by the daemon layer.
type AgentStatus struct {
	AgentID string `json:"agent_id,omitempty"`
	Name    string `json:"name"`
	State   string `json:"state,omitempty"`
}

// RuntimeModel is the runtime-scoped model catalog projected to the
// control plane. Model IDs are opaque to the daemon except for the local
// provider config source that reported them.
type RuntimeModel struct {
	ModelID   string `json:"model_id"`
	Label     string `json:"label"`
	IsDefault bool   `json:"is_default"`
}

// Status is the synchronous Status(ctx) snapshot.
type Status struct {
	RuntimeID       string         `json:"runtime_id"`
	StartedAt       time.Time      `json:"started_at"`
	UptimeSeconds   int64          `json:"uptime_seconds"`
	Health          string         `json:"health"`
	Owner           string         `json:"owner,omitempty"`
	DeviceName      string         `json:"device_name,omitempty"`
	Agents          []AgentStatus  `json:"agents,omitempty"`
	Models          []RuntimeModel `json:"models,omitempty"`
	Capabilities    []Capability   `json:"capabilities"`
	MaxConcurrent   int            `json:"max_concurrent"`
	RunningSessions int            `json:"running_sessions"`
	RunningTasks    []TaskStatus   `json:"running_tasks"`
}

// Heartbeat is the publish-ready payload for ControlPlane.Heartbeat.
type Heartbeat struct {
	RuntimeID      string   `json:"runtime_id"`
	UptimeSeconds  int64    `json:"uptime_seconds"`
	DeviceName     string   `json:"device_name,omitempty"`
	SlotLimit      int      `json:"slot_limit"`
	SlotsInUse     int      `json:"slots_in_use"`
	RunningTaskIDs []string `json:"running_task_ids"`
}
