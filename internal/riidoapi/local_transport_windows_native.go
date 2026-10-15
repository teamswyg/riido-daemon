//go:build windows

package riidoapi

import (
	"errors"
	"syscall"
	"unsafe"
)

const (
	namedPipeAccessDuplex = 0x00000003
	namedPipeTypeByte     = 0x00000000
	namedPipeReadModeByte = 0x00000000
	namedPipeWait         = 0x00000000
	namedPipeRejectRemote = 0x00000008
	namedPipeInstances    = 255
	namedPipeBufferSize   = 64 * 1024

	errorPipeBusy      syscall.Errno = 231
	errorFileNotFound  syscall.Errno = 2
	errorBrokenPipe    syscall.Errno = 109
	errorNoData        syscall.Errno = 232
	errorPipeNotConn   syscall.Errno = 233
	errorPipeConnected syscall.Errno = 535
)

var (
	kernel32ProcConnectNamedPipe = syscall.NewLazyDLL("kernel32.dll").NewProc("ConnectNamedPipe")
	kernel32ProcCreateNamedPipe  = syscall.NewLazyDLL("kernel32.dll").NewProc("CreateNamedPipeW")
)

func createNamedPipe(path string) (syscall.Handle, error) {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return syscall.InvalidHandle, err
	}
	handle, _, callErr := kernel32ProcCreateNamedPipe.Call(
		uintptr(unsafe.Pointer(name)),
		uintptr(namedPipeAccessDuplex),
		uintptr(namedPipeTypeByte|namedPipeReadModeByte|namedPipeWait|namedPipeRejectRemote),
		uintptr(namedPipeInstances),
		uintptr(namedPipeBufferSize),
		uintptr(namedPipeBufferSize),
		0,
		0,
	)
	if syscall.Handle(handle) == syscall.InvalidHandle {
		return syscall.InvalidHandle, callErr
	}
	return syscall.Handle(handle), nil
}

func connectNamedPipe(handle syscall.Handle) (bool, error) {
	ok, _, callErr := kernel32ProcConnectNamedPipe.Call(uintptr(handle), 0)
	if ok != 0 {
		return true, nil
	}
	var errno syscall.Errno
	isErrno := errors.As(callErr, &errno)
	if isErrno && errno == errorPipeConnected {
		return true, nil
	}
	return false, callErr
}
