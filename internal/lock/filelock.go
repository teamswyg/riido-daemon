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
	file *os.File
	path string
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
			return &FileLock{file: file, path: path}, nil
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

// Release releases the advisory lock and closes the underlying file.
func (l *FileLock) Release() error {
	if l == nil || l.file == nil {
		return nil
	}
	err := unlockFile(l.file, l.path)
	closeErr := l.file.Close()
	path := l.path
	l.file = nil
	l.path = ""
	if err != nil {
		return fmt.Errorf("lock: release file lock: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("lock: close file lock: %w", closeErr)
	}
	if err := cleanupLockFile(path); err != nil {
		return fmt.Errorf("lock: cleanup file lock: %w", err)
	}
	return nil
}

// WithFile runs fn while holding an exclusive advisory file lock.
func WithFile(ctx context.Context, path string, fn func() error) error {
	lock, err := AcquireFile(ctx, path)
	if err != nil {
		return err
	}
	defer lock.Release()
	return fn()
}
