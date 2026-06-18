package riidoapi

import "github.com/teamswyg/riido-daemon/internal/taskdb"

func statusFromDB(config Config, db taskdb.TaskDB) Status {
	return Status{
		SchemaVersion:       StatusSchemaVersion,
		Transport:           string(normalizeLocalTransport(config.Transport)),
		SocketPath:          config.SocketPath,
		TaskDBPath:          config.TaskDBPath,
		TaskDBSchemaVersion: db.SchemaVersion,
		TaskCount:           len(db.Tasks),
		TransitionCount:     len(db.Transitions),
		EvidenceCount:       len(db.Evidence),
		CommandReceiptCount: len(db.CommandReceipts),
		DiagnosticCount:     len(db.Diagnostics),
		UpdatedAt:           db.UpdatedAt,
	}
}
