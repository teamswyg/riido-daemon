package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

// RuntimeRegistry is the task DB source sidecar written next to the
// riido-task-db.v1 file. It lets local GUI/Zed integrations inspect
// runtime registration and heartbeat without reaching into daemon memory.
type RuntimeRegistry struct {
	SchemaVersion string                           `json:"schema_version"`
	TaskDBPath    string                           `json:"task_db_path"`
	UpdatedAt     time.Time                        `json:"updated_at"`
	Runtimes      []controlplane.RegisteredRuntime `json:"runtimes"`
}
