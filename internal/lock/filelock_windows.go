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

func startLockMaintenance(path string) func() {
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(fileLockClaimStaleAfter / 3)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = refreshLockClaim(path)
			case <-stop:
				return
			}
		}
	}()
	return func() {
		close(stop)
		<-done
	}
}

func writeNewLockClaim(path string) error {
	claimPath := lockClaimPath(path)
	claim, err := os.OpenFile(claimPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if err := writeLockClaim(claim, path, time.Now()); err != nil {
		_ = claim.Close()
		_ = os.Remove(claimPath)
		return err
	}
	return claim.Close()
}

func refreshLockClaim(path string) error {
	claim, err := os.OpenFile(lockClaimPath(path), os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if err := writeLockClaim(claim, path, time.Now()); err != nil {
		_ = claim.Close()
		return err
	}
	return claim.Close()
}

func writeLockClaim(file *os.File, path string, now time.Time) error {
	body, err := encodeFileLockClaim(newFileLockClaim(path, now))
	if err != nil {
		return err
	}
	_, err = file.Write(body)
	return err
}
