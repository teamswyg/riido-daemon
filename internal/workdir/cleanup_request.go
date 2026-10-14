package workdir

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"
)

func cleanupRequestBounds(root string, req CleanupRequest) (string, time.Time, time.Time, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", time.Time{}, time.Time{}, errors.New("workdir: cleanup root is required")
	}
	cutoff := req.ArchivedBefore
	if cutoff.IsZero() {
		return "", time.Time{}, time.Time{}, errors.New("workdir: cleanup ArchivedBefore is required")
	}
	removedAt := req.RemovedAt
	if removedAt.IsZero() {
		removedAt = time.Now()
	}
	return root, cutoff.UTC(), removedAt.UTC(), nil
}

func ensureCleanupRoot(root string) (bool, error) {
	info, err := os.Stat(root)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("workdir: stat cleanup root: %w", err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("workdir: cleanup root is not a directory: %s", root)
	}
	return true, nil
}
