package controlplane

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// FileClaimRecordSchemaVersion is the local file queue claim receipt schema.
const FileClaimRecordSchemaVersion = "riido-file-queue-claim.v1"

// FileClaimRecord is written under queue/claims/ when FileQueueSource
// atomically claims a top-level task JSON file.
type FileClaimRecord struct {
	SchemaVersion string             `json:"schema_version"`
	TaskID        string             `json:"task_id"`
	RuntimeID     string             `json:"runtime_id"`
	SourceFile    string             `json:"source_file"`
	ClaimedAt     time.Time          `json:"claimed_at"`
	Task          bridge.TaskRequest `json:"task"`
}
