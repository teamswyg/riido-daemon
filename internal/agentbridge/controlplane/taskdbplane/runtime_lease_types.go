package taskdbplane

import "time"

// RuntimeLeaseRegistry is the task DB source sidecar that records the
// latest local C9 fencing token per task.
type RuntimeLeaseRegistry struct {
	SchemaVersion string               `json:"schema_version"`
	TaskDBPath    string               `json:"task_db_path"`
	UpdatedAt     time.Time            `json:"updated_at"`
	Leases        []RuntimeLeaseRecord `json:"leases"`
}

type RuntimeLeaseRecord struct {
	LeaseID               string     `json:"lease_id"`
	TaskID                string     `json:"task_id"`
	RuntimeID             string     `json:"runtime_id"`
	CapabilityFingerprint string     `json:"capability_fingerprint,omitempty"`
	ClaimedAt             time.Time  `json:"claimed_at"`
	LeaseUntil            time.Time  `json:"lease_until"`
	FencingToken          int64      `json:"fencing_token"`
	ReleasedAt            *time.Time `json:"released_at,omitempty"`
}
