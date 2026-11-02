//go:build windows

package securecredential

import (
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

type windowsLockAPI struct {
	createRelative                             *syscall.LazyProc
	lock, release, driveType                   *syscall.LazyProc
	convertSD, securityInfo, descriptorControl *syscall.LazyProc
	validACL, getACE, validSID, equalSID       *syscall.LazyProc
}

var windowsLockAPICache struct {
	once sync.Once
	api  *windowsLockAPI
	err  error
}

func loadWindowsLockAPI() (*windowsLockAPI, error) {
	windowsLockAPICache.once.Do(func() {
		windowsLockAPICache.api, windowsLockAPICache.err = newWindowsLockAPI()
	})
	return windowsLockAPICache.api, windowsLockAPICache.err
}

func newWindowsLockAPI() (*windowsLockAPI, error) {
	// kernel32 is a KnownDLL. Other modules use the OS system directory, never
	// the working directory, PATH or an environment-provided SystemRoot.
	kernel := syscall.NewLazyDLL("kernel32.dll")
	getSystemDirectory := kernel.NewProc("GetSystemDirectoryW")
	if err := getSystemDirectory.Find(); err != nil {
		return nil, ErrUnavailable
	}
	buffer := make([]uint16, 32768)
	n, _, _ := getSystemDirectory.Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if n == 0 || n >= uintptr(len(buffer)) {
		return nil, ErrUnavailable
	}
	system := syscall.UTF16ToString(buffer[:n])
	if !filepath.IsAbs(system) {
		return nil, ErrUnavailable
	}
	security := syscall.NewLazyDLL(filepath.Join(system, "advapi32.dll"))
	ntdll := syscall.NewLazyDLL(filepath.Join(system, "ntdll.dll"))
	api := &windowsLockAPI{
		createRelative: ntdll.NewProc("NtCreateFile"),
		lock:           kernel.NewProc("LockFileEx"), release: kernel.NewProc("UnlockFileEx"),
		driveType:         kernel.NewProc("GetDriveTypeW"),
		convertSD:         security.NewProc("ConvertStringSecurityDescriptorToSecurityDescriptorW"),
		securityInfo:      security.NewProc("GetSecurityInfo"),
		descriptorControl: security.NewProc("GetSecurityDescriptorControl"),
		validACL:          security.NewProc("IsValidAcl"), getACE: security.NewProc("GetAce"),
		validSID: security.NewProc("IsValidSid"), equalSID: security.NewProc("EqualSid"),
	}
	for _, proc := range []*syscall.LazyProc{
		api.lock, api.release, api.driveType, api.createRelative,
		api.convertSD, api.securityInfo, api.descriptorControl, api.validACL,
		api.getACE, api.validSID, api.equalSID,
	} {
		if err := proc.Find(); err != nil {
			return nil, ErrUnavailable
		}
	}
	return api, nil
}

func (a *windowsLockAPI) tryLock(handle syscall.Handle) error {
	var offset syscall.Overlapped
	ok, _, err := a.lock.Call(uintptr(handle), 3, 0, 1, 0, uintptr(unsafe.Pointer(&offset)))
	if ok == 0 {
		return err
	}
	return nil
}

func (a *windowsLockAPI) unlock(handle syscall.Handle) {
	var offset syscall.Overlapped
	_, _, _ = a.release.Call(uintptr(handle), 0, 1, 0, uintptr(unsafe.Pointer(&offset)))
}
