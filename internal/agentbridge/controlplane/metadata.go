package controlplane

import "github.com/teamswyg/riido-contracts/metadatakeys"

const (
	MetadataRuntimeLeaseID               = string(metadatakeys.RuntimeLeaseID)
	MetadataRuntimeFencingToken          = string(metadatakeys.RuntimeFencingToken)
	MetadataRuntimeCapabilityFingerprint = string(metadatakeys.RuntimeCapabilityFingerprint)
	// MetadataTaskID preserves the logical Riido task id when a source uses
	// TaskRequest.ID as a run/execution id. Sources that already use
	// TaskRequest.ID as the logical task id do not need to set it.
	MetadataTaskID = string(metadatakeys.TaskID)
)
