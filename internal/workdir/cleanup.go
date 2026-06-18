package workdir

import (
	"context"
	"fmt"
	"path/filepath"
)

// CleanupArchivedBefore deletes run roots whose archive manifest is
// keep-in-place and older than req.ArchivedBefore. Runs without
// archive.json are considered active or dirty and are never removed.
func (a *FSAdapter) CleanupArchivedBefore(ctx context.Context, req CleanupRequest) (CleanupResult, error) {
	var result CleanupResult
	root, cutoff, removedAt, err := cleanupRequestBounds(a.root, req)
	if err != nil {
		return result, err
	}
	exists, err := ensureCleanupRoot(root)
	if err != nil {
		return result, err
	}
	if !exists {
		return result, nil
	}
	err = filepath.WalkDir(root, cleanupArchiveWalk(ctx, root, cutoff, removedAt, &result))
	if err != nil {
		return result, fmt.Errorf("workdir: cleanup archived runs: %w", err)
	}
	return result, nil
}
