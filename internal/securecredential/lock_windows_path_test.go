//go:build windows

package securecredential

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestWindowsCredentialLockRejectsHardlinksAndDirectories(t *testing.T) {
	for _, kind := range []string{"hardlink", "directory", "inherited-directory"} {
		t.Run(kind, func(t *testing.T) {
			s := windowsLockTestStore(t)
			if kind == "inherited-directory" {
				if err := os.Mkdir(s.LockDir, 0o700); err != nil {
					t.Fatal(err)
				}
			} else {
				unlock, err := s.TryLock(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				unlock()
				path := filepath.Join(s.LockDir, s.Key+".lock")
				if kind == "hardlink" {
					if err := os.Link(path, filepath.Join(s.LockDir, "other")); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					if err := os.Mkdir(path, 0o700); err != nil {
						t.Fatal(err)
					}
				}
			}
			if unlock, err := s.TryLock(t.Context()); unlock != nil || !errors.Is(err, ErrUnavailable) {
				t.Fatal("unsafe object accepted", err)
			}
		})
	}
}

func TestWindowsCredentialLockPinsAncestorsAndFile(t *testing.T) {
	base := filepath.Join(t.TempDir(), "ancestor")
	s := &Store{Key: "account-0123456789", LockDir: filepath.Join(base, "locks")}
	unlock, err := s.TryLock(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	file := filepath.Join(s.LockDir, s.Key+".lock")
	for _, path := range []string{base, s.LockDir, file} {
		if err := os.Rename(path, path+"-moved"); err == nil {
			_ = os.Rename(path+"-moved", path)
			t.Fatal("locked path could be replaced")
		}
	}
	if err := os.Remove(file); err == nil {
		t.Fatal("locked file could be deleted")
	}
	// Sharing denies GENERIC_WRITE handles. It does not exclude attribute-only
	// changes; the separate regression verifies no-reparse relative child opens.
	for _, path := range []string{base, s.LockDir} {
		name, err := syscall.UTF16PtrFromString(path)
		if err != nil {
			t.Fatal(err)
		}
		handle, err := syscall.CreateFile(name, syscall.GENERIC_WRITE,
			syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
			nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OPEN_REPARSE_POINT, 0)
		if err == nil {
			_ = syscall.CloseHandle(handle)
			t.Fatal("ancestor could be opened for reparse conversion")
		}
		if !errors.Is(err, syscall.Errno(32)) { // ERROR_SHARING_VIOLATION
			t.Fatal("ancestor denial did not prove handle pinning", err)
		}
	}
	unlock()
	if err := os.Rename(base, base+"-moved"); err != nil {
		t.Fatal("released ancestor retained a handle", err)
	}
	if err := os.Rename(base+"-moved", base); err != nil {
		t.Fatal(err)
	}
}
