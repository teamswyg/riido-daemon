//go:build windows

package main

import (
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func loadWindowsAuthorizationShell() (authorizationShell, error) {
	dir, err := authorizationSystemDirectory()
	if err != nil {
		return authorizationShell{}, err
	}
	// GetSystemDirectoryW supplies the trusted OS path, not WINDIR or PATH.
	shell := syscall.NewLazyDLL(filepath.Join(dir, "shell32.dll")).NewProc("ShellExecuteExW")
	ole := syscall.NewLazyDLL(filepath.Join(dir, "ole32.dll"))
	initialize, uninitialize := ole.NewProc("CoInitializeEx"), ole.NewProc("CoUninitialize")
	for _, proc := range []*syscall.LazyProc{shell, initialize, uninitialize} {
		if proc.Find() != nil {
			return authorizationShell{}, standalonecli.ErrBrowserUnavailable
		}
	}
	return authorizationShell{
		initialize: func() bool {
			hresult, _, _ := initialize.Call(0, 0x2|0x4) // STA; disable OLE1 DDE.
			return uint32(hresult) == 0 || uint32(hresult) == 1
		},
		uninitialize: func() { _, _, _ = uninitialize.Call() },
		execute: func(info *authorizationShellInfo) bool {
			ok, _, _ := shell.Call(uintptr(unsafe.Pointer(info)))
			return ok != 0
		},
	}, nil
}

func authorizationSystemDirectory() (string, error) {
	// kernel32 is an OS KnownDLL. Other libraries are loaded by absolute path.
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GetSystemDirectoryW")
	if proc.Find() != nil {
		return "", standalonecli.ErrBrowserUnavailable
	}
	buffer := make([]uint16, 32768)
	n, _, _ := proc.Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if n == 0 || n >= uintptr(len(buffer)) {
		return "", standalonecli.ErrBrowserUnavailable
	}
	dir := syscall.UTF16ToString(buffer[:n])
	if !filepath.IsAbs(dir) {
		return "", standalonecli.ErrBrowserUnavailable
	}
	return dir, nil
}
