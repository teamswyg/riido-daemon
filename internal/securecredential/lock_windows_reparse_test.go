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

func TestWindowsCredentialLockRejectsFileSymlink(t *testing.T) {
	s := windowsLockTestStore(t)
	_, security := windowsLockTestSecurity(t)
	windowsLockTestCreateDirectory(t, s.LockDir, security)
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte("unchanged"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := os.Symlink(target, filepath.Join(s.LockDir, s.Key+".lock"))
	if errors.Is(err, syscall.Errno(1314)) { // ERROR_PRIVILEGE_NOT_HELD
		t.Skip("host lacks symlink privilege; directory junction checks still run")
	}
	if err != nil {
		t.Fatal(err)
	}
	if unlock, err := s.TryLock(t.Context()); unlock != nil || !errors.Is(err, ErrUnavailable) {
		t.Fatal("file symlink accepted", err)
	}
	if contents, err := os.ReadFile(target); err != nil || string(contents) != "unchanged" {
		t.Fatal("rejected symlink target was changed", err)
	}
}

func TestWindowsCredentialLockRejectsDirectoryReparsePoints(t *testing.T) {
	for _, ancestor := range []bool{false, true} {
		name := "lock-directory"
		if ancestor {
			name = "ancestor-directory"
		}
		t.Run(name, func(t *testing.T) {
			base := t.TempDir()
			target := filepath.Join(base, "target")
			if err := os.Mkdir(target, 0o700); err != nil {
				t.Fatal(err)
			}
			junction := filepath.Join(base, "junction")
			windowsLockTestJunction(t, junction, target)
			dir := junction
			if ancestor {
				dir = filepath.Join(junction, "locks")
			}
			s := &Store{Key: "account-0123456789", LockDir: dir}
			if unlock, err := s.TryLock(t.Context()); unlock != nil || !errors.Is(err, ErrUnavailable) {
				t.Fatal("junction accepted", err)
			}
			if entries, err := os.ReadDir(target); err != nil || len(entries) != 0 {
				t.Fatal("rejected junction target was changed", err)
			}
		})
	}
}

func windowsLockTestJunction(t *testing.T, path, target string) {
	t.Helper()
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := windowsLockTestConvertDirectory(t, path, target, syscall.GENERIC_WRITE); err != nil {
		t.Fatal("cannot create isolated junction fixture", err)
	}
}

func windowsLockTestConvertDirectory(t *testing.T, path, target string, access uint32) error {
	t.Helper()
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := syscall.CreateFile(name, access,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE, nil,
		syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return err
	}
	defer func() { _ = syscall.CloseHandle(handle) }()
	buffer := windowsLockTestReparseBuffer(t, target)
	var returned uint32
	return syscall.DeviceIoControl(handle, 0x900a4, &buffer[0], uint32(len(buffer)), nil, 0, &returned, nil)
}

func windowsLockTestReparseBuffer(t *testing.T, target string) []byte {
	t.Helper()
	substitute, err := syscall.UTF16FromString(`\??\` + target)
	if err != nil {
		t.Fatal(err)
	}
	printName, err := syscall.UTF16FromString(target)
	if err != nil {
		t.Fatal(err)
	}
	// A mount-point reparse buffer creates a junction without symlink privilege.
	// Only this test's fresh directory and target are touched.
	buffer := make([]byte, 16+2*(len(substitute)+len(printName)))
	binary.LittleEndian.PutUint32(buffer[0:4], 0xa0000003)
	binary.LittleEndian.PutUint16(buffer[4:6], uint16(len(buffer)-8))
	binary.LittleEndian.PutUint16(buffer[10:12], uint16(2*(len(substitute)-1)))
	binary.LittleEndian.PutUint16(buffer[12:14], uint16(2*len(substitute)))
	binary.LittleEndian.PutUint16(buffer[14:16], uint16(2*(len(printName)-1)))
	for i, value := range append(substitute, printName...) {
		binary.LittleEndian.PutUint16(buffer[16+i*2:18+i*2], value)
	}
	return buffer
}
