// Package taskdbplane adapts riido-task-db.v1 into the agentbridge
// control-plane ports.
//
// It is intentionally outside the core controlplane package: the
// port definitions stay independent from project persistence, while
// this adapter is allowed to translate taskdb.TaskRecord rows into
// bridge.TaskRequest values and report guarded TaskState transitions.
package taskdbplane

import "time"

const (
	RuntimeRegistrySchemaVersion      = "riido-runtime-registry.v1"
	RuntimeLeaseRegistrySchemaVersion = "riido-runtime-lease-registry.v1"

	sourceName         = "riido.agentbridge.taskdb"
	metadataTaskDB     = "task_db_path"
	metadataDocument   = "source_document_path"
	commandIDPrefix    = "command:riido.agentbridge.taskdb:"
	defaultActor       = "daemon"
	defaultClaimReason = "runtime claimed queued task DB row"

	defaultRuntimeLeaseTTL = 30 * time.Second
)
