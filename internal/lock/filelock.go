// Package lock owns C9 local locking primitives.
//
// It provides infrastructure mechanics only. Domain lease meaning stays in
// internal/scheduling, and task DB adapters decide when to acquire a lock.
package lock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileLock is an exclusive advisory lock backed by the host OS lock primitive.
type FileLock struct {
	file            *os.File
	path            string
	stopMaintenance func()
}

// AcquireFile waits until an exclusive lock can be acquired or ctx is done.
func AcquireFile(ctx context.Context, path string) (*FileLock, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("lock: empty path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("lock: create lock directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("lock: open file lock: %w", err)
	}
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		err := tryLockFile(file, path)
		if err == nil {
			return &FileLock{file: file, path: path, stopMaintenance: startLockMaintenance(path)}, nil
		}
		if !isLockBusy(err) {
			_ = file.Close()
			return nil, fmt.Errorf("lock: acquire file lock: %w", err)
		}
		select {
		case <-ctx.Done():
			_ = file.Close()
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
