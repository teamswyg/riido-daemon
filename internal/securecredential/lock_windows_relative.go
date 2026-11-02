//go:build windows

package securecredential

import (
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

type windowsLockUnicodeString struct {
	length, maximumLength uint16
	buffer                *uint16
}

type windowsLockObjectAttributes struct {
	length             uint32
	root               syscall.Handle
	name               *windowsLockUnicodeString
	flags              uint32
	securityDescriptor unsafe.Pointer
	securityQoS        unsafe.Pointer
}

type windowsLockIOStatus struct {
	status, information uintptr
}

func (a *windowsLockAPI) openRelative(security *windowsLockSecurity, parent syscall.Handle, name string, directory bool) (syscall.Handle, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `\/:`) || !filepath.IsLocal(name) || strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
		return syscall.InvalidHandle, ErrUnavailable
	}
	buffer, err := syscall.UTF16FromString(name)
	if err != nil || len(buffer) > 32767 {
		return syscall.InvalidHandle, ErrUnavailable
	}
	text := windowsLockUnicodeString{length: uint16(2 * (len(buffer) - 1)), maximumLength: uint16(2 * len(buffer)), buffer: &buffer[0]}
	attributes := windowsLockObjectAttributes{
		length: uint32(unsafe.Sizeof(windowsLockObjectAttributes{})), root: parent, name: &text,
		flags: 0x1040, securityDescriptor: security.descriptor, // OBJ_DONT_REPARSE | OBJ_CASE_INSENSITIVE; never OBJ_INHERIT.
	}
	access := uint32(syscall.GENERIC_READ | syscall.GENERIC_WRITE | windowsLockReadControl | syscall.SYNCHRONIZE)
	share := uint32(syscall.FILE_SHARE_READ | syscall.FILE_SHARE_WRITE)
	options := uint32(0x20 | 0x40) // FILE_SYNCHRONOUS_IO_NONALERT | FILE_NON_DIRECTORY_FILE.
	if directory {
		access = windowsLockReadControl | windowsLockReadAttributes | syscall.SYNCHRONIZE | 1 // FILE_LIST_DIRECTORY.
		share = syscall.FILE_SHARE_READ
		options = 0x20 | 1 // FILE_SYNCHRONOUS_IO_NONALERT | FILE_DIRECTORY_FILE.
	}
	var handle syscall.Handle
	var status windowsLockIOStatus
	// Every name has exactly one component and is resolved from the verified
	// parent handle. An attribute-only junction mutation cannot redirect a new
	// absolute-path lookup. OBJ_DONT_REPARSE rejects reparse traversal in the OS.
	result, _, _ := a.createRelative.Call(uintptr(unsafe.Pointer(&handle)), uintptr(access),
		uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&status)), 0,
		syscall.FILE_ATTRIBUTE_NORMAL, uintptr(share), 3, uintptr(options), 0, 0) // FILE_OPEN_IF.
	if result != 0 || handle == 0 || handle == syscall.InvalidHandle {
		if handle != 0 && handle != syscall.InvalidHandle {
			_ = syscall.CloseHandle(handle)
		}
		return syscall.InvalidHandle, ErrUnavailable
	}
	return handle, nil
}
