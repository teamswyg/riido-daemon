// Package taskdb owns the public daemon's local riido-task-db.v1 persistence
// model and guarded mutation rules.
//
// It does not own workspace projection, mwsd synchronization, local IPC, or
// provider execution. Those contexts feed task rows into this package or consume
// its receipts through explicit adapters.
package taskdb

const (
	TaskDBSchemaVersion       = "riido-task-db.v1"
	TaskCommandReplayPolicyV1 = "command-id-idempotent-replay.v1"
	TaskEvidenceValidationV1  = "deterministic-command-exit-code.v1"
)
