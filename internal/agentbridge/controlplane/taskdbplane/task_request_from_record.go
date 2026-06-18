package taskdbplane

import (
	"strconv"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/supervisor"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func taskRequestFromRecord(path string, record taskdb.TaskRecord, provider, prompt string, lease RuntimeLeaseRecord) bridge.TaskRequest {
	meta := map[string]string{
		metadataTaskDB: path,
	}
	if record.ProjectID != "" {
		meta[supervisor.MetadataWorkspaceID] = record.ProjectID
	}
	if record.SourceDocumentPath != "" {
		meta[metadataDocument] = record.SourceDocumentPath
	}
	addLeaseMetadata(meta, lease)
	return bridge.TaskRequest{
		ID:       record.ID,
		Provider: bridge.Provider(provider),
		Prompt:   prompt,
		Metadata: meta,
	}
}

func addLeaseMetadata(meta map[string]string, lease RuntimeLeaseRecord) {
	if lease.LeaseID == "" {
		return
	}
	meta[controlplane.MetadataRuntimeLeaseID] = lease.LeaseID
	meta[controlplane.MetadataRuntimeFencingToken] = strconv.FormatInt(lease.FencingToken, 10)
	if lease.CapabilityFingerprint != "" {
		meta[controlplane.MetadataRuntimeCapabilityFingerprint] = lease.CapabilityFingerprint
	}
}
