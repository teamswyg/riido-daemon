//go:build windows

package lock

import (
	"os"
	"time"
)

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
