//go:build !windows

package lock

import (
	"errors"
	"os"
	"syscall"
)

func tryLockFile(file *os.File, _ string) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func isLockBusy(err error) bool {
	return errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN)
}

func unlockFile(file *os.File, _ string) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

func cleanupLockFile(string) error {
	return nil
}
