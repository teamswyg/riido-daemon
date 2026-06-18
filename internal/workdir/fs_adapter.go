package workdir

import "context"

// Archiver is an optional port implemented by adapters that can record
// terminal workspace lifecycle state.
type Archiver interface {
	Archive(Workspace, ArchiveRequest) (ArchiveRecord, error)
}

// Cleaner is an optional port implemented by adapters that can delete
// archived run roots once an operator-supplied retention cutoff expires.
type Cleaner interface {
	CleanupArchivedBefore(context.Context, CleanupRequest) (CleanupResult, error)
}

// FSAdapter is the filesystem implementation rooted at a single path.
type FSAdapter struct {
	root string
}

// NewFSAdapter constructs an adapter rooted at root. The root is created
// lazily by Prepare.
func NewFSAdapter(root string) *FSAdapter { return &FSAdapter{root: root} }
