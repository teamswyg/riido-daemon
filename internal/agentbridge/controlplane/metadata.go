package controlplane

const (
	MetadataRuntimeLeaseID               = "runtime_lease_id"
	MetadataRuntimeFencingToken          = "runtime_fencing_token"
	MetadataRuntimeCapabilityFingerprint = "runtime_capability_fingerprint"
	// MetadataTaskID preserves the logical Riido task id when a source uses
	// TaskRequest.ID as a run/execution id. Sources that already use
	// TaskRequest.ID as the logical task id do not need to set it.
	MetadataTaskID = "task_id"
)
