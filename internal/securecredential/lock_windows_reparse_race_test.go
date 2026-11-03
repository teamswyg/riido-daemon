//go:build windows

package securecredential

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestWindowsCredentialLockAttributeOnlyReparseDoesNotEscape(t *testing.T) {
	for _, directory := range []bool{false, true} {
		kind := "file"
		if directory {
			kind = "child-directory"
		}
		t.Run(kind, func(t *testing.T) {
			base := t.TempDir()
			api, security := windowsLockTestSecurity(t)
			target := filepath.Join(base, "outside")
			windowsLockTestCreateDirectory(t, target, security)
			control := filepath.Join(base, "control")
			windowsLockTestCreateDirectory(t, control, security)
			// Prove the fixture can convert an unpinned empty directory using
			// FILE_WRITE_ATTRIBUTES alone, before relying on any negative result.
			if err := windowsLockTestConvertDirectory(t, control, target, 0x100); err != nil {
				t.Fatal("attribute-only positive control did not convert", err)
			}
			t.Cleanup(func() { windowsLockTestRemoveReparse(t, control) })
			windowsLockTestRequireReparse(t, control)
			victim := filepath.Join(base, "victim")
			handles, _, err := openWindowsLockDirectories(t.Context(), api, security, victim)
			if err != nil {
				t.Fatal(err)
			}
			defer closeWindowsLockHandles(handles)
			parent := handles[len(handles)-1]
			// The parent was validated above and is still empty. This is the
			// exact validation-to-child-open boundary; no lock file exists yet.
			if err := windowsLockTestConvertDirectory(t, victim, target, 0x100); err != nil {
				t.Fatal("attribute-only mutation of the pinned fixture failed", err)
			}
			t.Cleanup(func() { windowsLockTestRemoveReparse(t, victim) })
			windowsLockTestRequireReparse(t, victim)
			// Exercise the kernel primitive directly, without the caller's
			// parent recheck, to cover a mutation after that last recheck too.
			handle, err := api.openRelative(security, parent, "kernel-child", directory)
			if err == nil {
				_ = syscall.CloseHandle(handle)
			} else if !errors.Is(err, ErrUnavailable) {
				t.Fatal("relative open returned an unexpected error", err)
			}
			if entries, err := os.ReadDir(target); err != nil || len(entries) != 0 {
				t.Fatal("relative open escaped through the mutated parent", err)
			}
			if directory {
				handle, err = openWindowsLockChildDirectory(api, security, parent, "verified-child")
			} else {
				handle, err = openWindowsLockFile(api, security, parent, "account-0123456789.lock")
			}
			if err == nil {
				_ = syscall.CloseHandle(handle)
				t.Fatal("changed parent was accepted")
			}
			if !errors.Is(err, ErrUnavailable) {
				t.Fatal("changed parent did not fail closed", err)
			}
		})
	}
}

func windowsLockTestRequireReparse(t *testing.T, path string) {
	t.Helper()
	handle := windowsLockTestOpenReparse(t, path, windowsLockReadAttributes)
	defer func() { _ = syscall.CloseHandle(handle) }()
	var info syscall.ByHandleFileInformation
	if err := syscall.GetFileInformationByHandle(handle, &info); err != nil || info.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT == 0 {
		t.Fatal("fixture did not become a reparse point", err)
	}
}

func windowsLockTestRemoveReparse(t *testing.T, path string) {
	t.Helper()
	handle := windowsLockTestOpenReparse(t, path, 0x100)
	defer func() { _ = syscall.CloseHandle(handle) }()
	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint32(buffer[:4], 0xa0000003)
	var returned uint32
	if err := syscall.DeviceIoControl(handle, 0x900ac, &buffer[0], uint32(len(buffer)), nil, 0, &returned, nil); err != nil {
		t.Error("failed to remove isolated test reparse data", err)
	}
}

func windowsLockTestOpenReparse(t *testing.T, path string, access uint32) syscall.Handle {
	t.Helper()
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := syscall.CreateFile(name, access,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE, nil,
		syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		t.Fatal(err)
	}
	return handle
}
