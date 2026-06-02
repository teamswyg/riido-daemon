//go:build !windows

package lock

import (
	"os"
	"syscall"
)

func tryLockFile(file *os.File, _ string) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func isLockBusy(err error) bool {
	return err == syscall.EWOULDBLOCK || err == syscall.EAGAIN
}

func unlockFile(file *os.File, _ string) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

func cleanupLockFile(string) error {
	return nil
}
