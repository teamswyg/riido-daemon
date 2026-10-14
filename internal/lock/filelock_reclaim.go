package lock

import (
	"errors"
	"os"
	"time"
)

func lockClaimPath(path string) string {
	return path + ".claim"
}

func reclaimStaleLockClaim(path string, now time.Time) error {
	claimPath := lockClaimPath(path)
	info, err := os.Stat(claimPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(claimPath)
	if err != nil {
		return err
	}
	if !fileLockClaimStale(raw, info.ModTime(), now) {
		return os.ErrExist
	}
	return os.Remove(claimPath)
}
