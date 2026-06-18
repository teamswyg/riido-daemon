//go:build windows

package riidoapi

import (
	"syscall"
	"unsafe"
)

const (
	namedPipeAccessDuplex = 0x00000003
	namedPipeTypeByte     = 0x00000000
	namedPipeReadModeByte = 0x00000000
	namedPipeWait         = 0x00000000
	namedPipeInstances    = 255
	namedPipeBufferSize   = 64 * 1024

	errorPipeBusy      syscall.Errno = 231
	errorPipeConnected syscall.Errno = 535
)

var (
	kernel32ProcConnectNamedPipe    = syscall.NewLazyDLL("kernel32.dll").NewProc("ConnectNamedPipe")
	kernel32ProcCreateNamedPipe     = syscall.NewLazyDLL("kernel32.dll").NewProc("CreateNamedPipeW")
	kernel32ProcDisconnectNamedPipe = syscall.NewLazyDLL("kernel32.dll").NewProc("DisconnectNamedPipe")
)

func createNamedPipe(path string) (syscall.Handle, error) {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return syscall.InvalidHandle, err
	}
	handle, _, callErr := kernel32ProcCreateNamedPipe.Call(
		uintptr(unsafe.Pointer(name)),
		uintptr(namedPipeAccessDuplex),
		uintptr(namedPipeTypeByte|namedPipeReadModeByte|namedPipeWait),
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
	errno, isErrno := callErr.(syscall.Errno)
	if isErrno && errno == errorPipeConnected {
		return true, nil
	}
	return false, callErr
}
