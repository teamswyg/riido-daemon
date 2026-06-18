package workdir

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func cleanupArchiveWalk(ctx context.Context, root string, cutoff, removedAt time.Time, result *CleanupResult) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, walkErr error) error {
		if err := cleanupWalkCanceled(ctx); err != nil {
			return err
		}
		if err := cleanupWalkError(walkErr); err != nil {
			return err
		}
		if d.IsDir() || filepath.Base(path) != "archive.json" {
			return nil
		}
		return cleanupArchivePath(path, root, cutoff, removedAt, result)
	}
}

func cleanupWalkCanceled(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func cleanupWalkError(walkErr error) error {
	if walkErr == nil || errors.Is(walkErr, fs.ErrNotExist) {
		return nil
	}
	return walkErr
}

func cleanupArchivePath(path, root string, cutoff, removedAt time.Time, result *CleanupResult) error {
	result.ScannedArchiveRecords++
	record, err := readArchiveRecord(path)
	if err != nil {
		return err
	}
	if !cleanupEligible(record, cutoff) {
		return nil
	}
	runRoot := filepath.Dir(path)
	if runRoot == root {
		return errors.New("workdir: refusing to remove cleanup root")
	}
	if err := os.RemoveAll(runRoot); err != nil {
		return fmt.Errorf("workdir: remove archived run %s: %w", runRoot, err)
	}
	result.Removed = append(result.Removed, CleanupRecord{
		RunRoot:   runRoot,
		Archive:   record,
		RemovedAt: removedAt,
	})
	return filepath.SkipDir
}
