package lock

import "fmt"

// Release releases the advisory lock and closes the underlying file.
func (l *FileLock) Release() error {
	if l == nil || l.file == nil {
		return nil
	}
	if l.stopMaintenance != nil {
		l.stopMaintenance()
		l.stopMaintenance = nil
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
