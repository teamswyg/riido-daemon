package saasplane

import (
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

type RuntimeModelRecord struct {
	ModelID   string `json:"model_id"`
	Label     string `json:"label"`
	IsDefault bool   `json:"is_default"`
}

type RuntimeSnapshotRecord struct {
	RuntimeID                 string               `json:"runtime_id"`
	Kind                      string               `json:"kind"`
	Availability              string               `json:"availability,omitempty"`
	DetectionState            string               `json:"detection_state,omitempty"`
	ProviderVersion           string               `json:"provider_version,omitempty"`
	RequiresExperimentalOptIn bool                 `json:"requires_experimental_opt_in,omitempty"`
	Models                    []RuntimeModelRecord `json:"models,omitempty"`
}

type DeviceRuntimeSnapshotSyncRequest struct {
	DaemonID          string                  `json:"daemon_id"`
	DeviceID          string                  `json:"device_id,omitempty"`
	DeviceDisplayName string                  `json:"device_display_name,omitempty"`
	Profile           string                  `json:"profile,omitempty"`
	AppVersion        string                  `json:"app_version,omitempty"`
	PID               int                     `json:"pid,omitempty"`
	UptimeSeconds     int64                   `json:"uptime_seconds,omitempty"`
	StartedAt         time.Time               `json:"started_at,omitzero"`
	Runtimes          []RuntimeSnapshotRecord `json:"runtimes"`
}

type AgentRuntimeBindingListResponse struct {
	SchemaVersion string                                   `json:"schema_version"`
	Bindings      []assignmentcontract.AgentRuntimeBinding `json:"bindings"`
}
