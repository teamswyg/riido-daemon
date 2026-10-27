//go:build darwin || linux

package securecredential

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// Lock covers the full load/refresh/save transaction across CLI processes.
// The file contains no token. It coordinates trusted CLI calls; it does not
// substitute for OS user-presence protection against a malicious OS peer.
func (s *Store) Lock(ctx context.Context) (func(), error) {
	if !validKey.MatchString(s.Key) || !filepath.IsAbs(s.LockDir) {
		return nil, ErrUnavailable
	}
	if err := os.MkdirAll(s.LockDir, 0o700); err != nil {
		return nil, ErrUnavailable
	}
	info, err := os.Lstat(s.LockDir)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0o077 != 0 || !owned(info) {
		return nil, ErrUnavailable
	}
	fd, err := syscall.Open(filepath.Join(s.LockDir, s.Key+".lock"), syscall.O_CREAT|syscall.O_RDWR|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0o600)
	if err != nil {
		return nil, ErrUnavailable
	}
	file := os.NewFile(uintptr(fd), "credential-lock")
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || !owned(info) {
		_ = file.Close()
		return nil, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	for {
		if err := ctx.Err(); err != nil {
			_ = file.Close()
			return nil, err
		}
		err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			_ = file.Close()
			return nil, ErrUnavailable
		}
		select {
		case <-ctx.Done():
			_ = file.Close()
			return nil, ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
	var once sync.Once
	return func() { once.Do(func() { _ = syscall.Flock(fd, syscall.LOCK_UN); _ = file.Close() }) }, nil
}

func owned(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Uid == uint32(os.Getuid()) && (info.IsDir() || stat.Nlink == 1)
}
