//go:build windows

package lock

import (
	"errors"
	"os"
)

func tryLockFile(_ *os.File, path string) error {
	claim, err := os.OpenFile(path+".claim", os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	return claim.Close()
}

func isLockBusy(err error) bool {
	return errors.Is(err, os.ErrExist)
}

func unlockFile(_ *os.File, _ string) error {
	return nil
}

func cleanupLockFile(path string) error {
	err := os.Remove(path + ".claim")
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
