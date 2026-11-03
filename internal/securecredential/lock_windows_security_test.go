//go:build windows

package securecredential

import (
	"errors"
	"path/filepath"
	"syscall"
	"testing"
	"unsafe"
)

func windowsLockTestSecurity(t *testing.T) (*windowsLockAPI, *windowsLockSecurity) {
	t.Helper()
	api, err := loadWindowsLockAPI()
	if err != nil {
		t.Fatal(err)
	}
	security, err := newWindowsLockSecurity(api)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(security.close)
	return api, security
}

func windowsLockTestDescriptor(t *testing.T, api *windowsLockAPI, sddl string) *windowsLockSecurity {
	t.Helper()
	text, err := syscall.UTF16PtrFromString(sddl)
	if err != nil {
		t.Fatal(err)
	}
	security := &windowsLockSecurity{}
	ok, _, err := api.convertSD.Call(uintptr(unsafe.Pointer(text)), 1, uintptr(unsafe.Pointer(&security.descriptor)), 0)
	if ok == 0 || security.descriptor == nil {
		t.Fatal("cannot create isolated test descriptor", err)
	}
	t.Cleanup(security.close)
	return security
}

func windowsLockTestCreateDirectory(t *testing.T, path string, security *windowsLockSecurity) {
	t.Helper()
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.CreateDirectory(name, security.attributes()); err != nil {
		t.Fatal(err)
	}
}

func TestWindowsCredentialLockRejectsBroadAndNullDACL(t *testing.T) {
	for _, target := range []string{"directory", "file"} {
		for _, policy := range []string{"broad", "null"} {
			t.Run(target+"-"+policy, func(t *testing.T) {
				s := windowsLockTestStore(t)
				api, private := windowsLockTestSecurity(t)
				sid, err := private.owner.String()
				if err != nil {
					t.Fatal(err)
				}
				sddl := "O:" + sid + "D:P(A;;FA;;;" + sid + ")(A;;FR;;;WD)"
				if policy == "null" {
					sddl = "O:" + sid + "D:NO_ACCESS_CONTROL"
				}
				unsafeSecurity := windowsLockTestDescriptor(t, api, sddl)
				if target == "directory" {
					windowsLockTestCreateDirectory(t, s.LockDir, unsafeSecurity)
				} else {
					windowsLockTestCreateDirectory(t, s.LockDir, private)
					name, err := syscall.UTF16PtrFromString(filepath.Join(s.LockDir, s.Key+".lock"))
					if err != nil {
						t.Fatal(err)
					}
					handle, err := syscall.CreateFile(name, syscall.GENERIC_READ|syscall.GENERIC_WRITE, 0,
						unsafeSecurity.attributes(), syscall.CREATE_NEW, syscall.FILE_ATTRIBUTE_NORMAL, 0)
					if err != nil {
						t.Fatal(err)
					}
					_ = syscall.CloseHandle(handle)
				}
				if unlock, err := s.TryLock(t.Context()); unlock != nil || !errors.Is(err, ErrUnavailable) {
					t.Fatal("unsafe ACL accepted", err)
				}
			})
		}
	}
}

func TestWindowsCredentialLockChecksOwnerAndHandleInheritance(t *testing.T) {
	s := windowsLockTestStore(t)
	api, security := windowsLockTestSecurity(t)
	directories, _, err := openWindowsLockDirectories(t.Context(), api, security, s.LockDir)
	if err != nil {
		t.Fatal(err)
	}
	defer closeWindowsLockHandles(directories)
	handle, err := openWindowsLockFile(api, security, directories[len(directories)-1], s.Key+".lock")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = syscall.CloseHandle(handle) }()
	foreign, err := syscall.StringToSid("S-1-1-0")
	if err != nil {
		t.Fatal(err)
	}
	if api.privateObject(handle, foreign) || api.privateObject(directories[len(directories)-1], foreign) {
		t.Fatal("different owner was accepted")
	}
	getFlags := syscall.NewLazyDLL("kernel32.dll").NewProc("GetHandleInformation")
	if err := getFlags.Find(); err != nil {
		t.Fatal(err)
	}
	for _, object := range append(directories, handle) {
		var flags uint32
		ok, _, err := getFlags.Call(uintptr(object), uintptr(unsafe.Pointer(&flags)))
		if ok == 0 || flags&1 != 0 {
			t.Fatal("lock handle can be inherited", err)
		}
	}
}
