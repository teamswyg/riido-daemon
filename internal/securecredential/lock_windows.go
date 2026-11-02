//go:build windows

package securecredential

import (
	"context"
	"errors"
	"slices"
	"sync"
	"syscall"
	"time"
)

// Lock covers the complete credential transaction. It coordinates CLI calls;
// it does not supply OS user presence or make the Windows store supported.
func (s *Store) Lock(ctx context.Context) (func(), error) { return s.lockWindows(ctx, true) }

// TryLock rejects a concurrent operation instead of queuing another operation.
func (s *Store) TryLock(ctx context.Context) (func(), error) { return s.lockWindows(ctx, false) }

func (s *Store) lockWindows(ctx context.Context, wait bool) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !validKey.MatchString(s.Key) {
		return nil, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	api, err := loadWindowsLockAPI()
	if err != nil {
		return nil, ErrUnavailable
	}
	security, err := newWindowsLockSecurity(api)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer security.close()
	directories, _, err := openWindowsLockDirectories(ctx, api, security, s.LockDir)
	if err != nil {
		return nil, windowsLockError(ctx, err)
	}
	file, err := openWindowsLockFile(api, security, directories[len(directories)-1], s.Key+".lock")
	if err != nil {
		closeWindowsLockHandles(directories)
		return nil, windowsLockError(ctx, err)
	}
	closeAll := func() {
		_ = syscall.CloseHandle(file)
		closeWindowsLockHandles(directories)
	}
	for {
		if err := ctx.Err(); err != nil {
			closeAll()
			return nil, err
		}
		err = api.tryLock(file)
		if err == nil {
			if err := ctx.Err(); err != nil {
				api.unlock(file)
				closeAll()
				return nil, err
			}
			var once sync.Once
			return func() { once.Do(func() { api.unlock(file); closeAll() }) }, nil
		}
		if !errors.Is(err, syscall.Errno(33)) { // ERROR_LOCK_VIOLATION
			closeAll()
			return nil, windowsLockError(ctx, err)
		}
		if !wait {
			closeAll()
			return nil, windowsLockError(ctx, ErrBusy)
		}
		timer := time.NewTimer(20 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			closeAll()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func windowsLockError(ctx context.Context, err error) error {
	if cancelled := ctx.Err(); cancelled != nil {
		return cancelled
	}
	if errors.Is(err, ErrBusy) {
		return ErrBusy
	}
	return ErrUnavailable
}

func closeWindowsLockHandles(handles []syscall.Handle) {
	for _, handle := range slices.Backward(handles) {
		_ = syscall.CloseHandle(handle)
	}
}
