package controlplane

import "time"

// RuntimeRegistration is the payload the daemon hands to the control
// plane when announcing a local runtime.
type RuntimeModel struct {
	ModelID   string `json:"model_id"`
	Label     string `json:"label"`
	IsDefault bool   `json:"is_default"`
}

type RuntimeRegistration struct {
	DaemonID             string            `json:"daemon_id"`
	RuntimeID            string            `json:"runtime_id"`
	Provider             string            `json:"provider"`
	Executable           string            `json:"executable,omitempty"`
	Version              string            `json:"version,omitempty"`
	Capabilities         map[string]bool   `json:"capabilities,omitempty"`
	CapabilityAttributes map[string]string `json:"capability_attributes,omitempty"`
	DeviceName           string            `json:"device_name,omitempty"`
	Models               []RuntimeModel    `json:"models,omitempty"`
	StartedAt            time.Time         `json:"started_at"`
	UptimeSeconds        int64             `json:"uptime_seconds,omitempty"`
	SlotLimit            int               `json:"slot_limit,omitempty"`
	SlotsInUse           int               `json:"slots_in_use,omitempty"`
	RunningTaskIDs       []string          `json:"running_task_ids,omitempty"`
}

// RuntimeHeartbeat is the periodic liveness/capacity snapshot the
// daemon hands to a task source after registration.
type RuntimeHeartbeat struct {
	RuntimeID      string   `json:"runtime_id"`
	UptimeSeconds  int64    `json:"uptime_seconds,omitempty"`
	DeviceName     string   `json:"device_name,omitempty"`
	SlotLimit      int      `json:"slot_limit,omitempty"`
	SlotsInUse     int      `json:"slots_in_use,omitempty"`
	RunningTaskIDs []string `json:"running_task_ids,omitempty"`
}

// RegisteredRuntime is the control plane's view of a runtime, including
// the last heartbeat it observed.
type RegisteredRuntime struct {
	RuntimeRegistration
	LastHeartbeat time.Time `json:"last_heartbeat"`
}
