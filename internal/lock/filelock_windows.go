//go:build windows

package lock

import (
	"errors"
	"os"
	"time"
)

func tryLockFile(_ *os.File, path string) error {
	if err := writeNewLockClaim(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrExist) {
		return err
	}
	if err := reclaimStaleLockClaim(path, time.Now()); err != nil {
		return err
	}
	return writeNewLockClaim(path)
}

func isLockBusy(err error) bool {
	return errors.Is(err, os.ErrExist)
}

func unlockFile(_ *os.File, _ string) error {
	return nil
}

func cleanupLockFile(path string) error {
	err := os.Remove(lockClaimPath(path))
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
